package website

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/model"
	"crawlers/pkg/stream"
	"github.com/reugn/go-streams/extension"
	"github.com/reugn/go-streams/flow"
	"go.uber.org/zap"
)

type StreamStepDefinition[T, R, E, U any] struct {
	sourceStream        string
	sourceConsumerGroup string
	destinationStream   string
	convertFunc         flow.MapFunction[T, R]
	flowFlatMap         flow.FlatMap[E, U]
}

func RegisterStream(ctx context.Context) error {
	pr := NewTaskProcessor()
	params := stream.DefaultStreamTaskParams()

	//consume catalog page ColumnUrl
	if err := catalogPageStream(ctx, pr, params); err != nil {
		return err
	}

	if err := novelStream(ctx, pr, params); err != nil {
		return err
	}

	if err := chapterStream(ctx, pr, params); err != nil {
		return err
	}
	return nil
}

func LaunchSiteTasks(ctx context.Context, siteName string) (err error) {
	params := stream.GenStreamTaskParams(siteName)
	pr := GetSiteTaskProcessor(siteName)

	//consume catalog page ColumnUrl
	if err = catalogPageStream(ctx, pr, params); err != nil {
		return err
	}

	if err = novelStream(ctx, pr, params); err != nil {
		return err
	}

	if err = chapterStream(ctx, pr, params); err != nil {
		return
	}
	return
}

// 解析page url得到每一个novel的url
// from: catalogPage stream => novel stream
func catalogPageStream(ctx context.Context, pr TaskProcessor, params *stream.StreamTaskParams) error {
	source, err := stream.NewRedisStreamSource(context.Background(), base.GetSystem().RedisClient,
		params.CatalogPageStreamName, params.CatalogPageStreamConsumer)
	if err != nil {
		return err
	}

	err = base.GetSystem().TaskPool.Submit(func() {
		source.
			Via(flow.NewMap(pr.HandleCatalogPageTask, 1)).
			Via(flow.NewFlatMap(func(novelMsg []model.NovelTask) []model.NovelTask {
				return novelMsg
			}, uint(base.GetConfig().CrawlerSettings.CatalogPageTaskParallelism))).
			To(stream.NewRedisStreamSink(ctx, base.GetSystem().RedisClient,
				params.NovelPageStreamName))
	})
	if err != nil {
		zap.S().Error("failed to submit task", zap.Error(err))
		return err
	}
	return nil
}

// 处理每一个novel
func novelStream(ctx context.Context, pr TaskProcessor, params *stream.StreamTaskParams) error {
	source, err := stream.NewRedisStreamSource(context.Background(), base.GetSystem().RedisClient,
		params.NovelPageStreamName, params.NovelPageStreamConsumer)
	if err != nil {
		return err
	}

	//item url
	sink := stream.NewRedisStreamSink(ctx, base.GetSystem().RedisClient,
		params.ChapterPageStreamName)

	err = base.GetSystem().TaskPool.Submit(func() {
		source.
			Via(flow.NewMap(pr.HandleNovelTask, uint(base.GetConfig().CrawlerSettings.NovelTaskParallelism))).
			Via(flow.NewFlatMap(func(novelMsg []model.ChapterTask) []model.ChapterTask {
				return novelMsg
			}, 1)).
			To(sink)
	})
	if err != nil {
		zap.L().Error("failed to submit task", zap.Error(err))
		return err
	}
	return nil
}

// 处理每一个novel
func chapterStream(ctx context.Context, pr TaskProcessor, params *stream.StreamTaskParams) error {
	source, err := stream.NewRedisStreamSource(ctx, base.GetSystem().RedisClient,
		params.ChapterPageStreamName, params.ChapterPageStreamConsumer)
	if err != nil {
		return err
	}

	err = base.GetSystem().TaskPool.Submit(func() {
		source.
			Via(flow.NewMap(pr.HandleChapterTask, uint(base.GetConfig().CrawlerSettings.ChapterTaskParallelism))).
			To(extension.NewIgnoreSink())
	})
	if err != nil {
		zap.L().Error("failed to submit task", zap.Error(err))
		return err
	}
	return nil
}
