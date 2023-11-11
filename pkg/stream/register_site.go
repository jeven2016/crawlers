package stream

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/model/entity"
	"crawlers/pkg/processor"
	"github.com/jeven2016/mylibs/system"
	"github.com/reugn/go-streams/extension"
	"github.com/reugn/go-streams/flow"
	"go.uber.org/zap"
)

type DefaultSiteStreamImpl struct {
	pr     processor.TaskProcessor
	params *StreamTaskParams
}

// LaunchGlobalSiteStream launches a global sharing site stream
func LaunchGlobalSiteStream(ctx context.Context) error {
	return LaunchSiteStream(ctx, "")
}

// LaunchSiteStream launch separated streams for a site
func LaunchSiteStream(ctx context.Context, siteName string) (err error) {
	siteStream := &DefaultSiteStreamImpl{
		params: GenStreamTaskParams(siteName),
		pr:     processor.GetSiteTaskProcessor(siteName),
	}

	//consume catalog page task message
	if err = siteStream.catalogPageStream(ctx); err != nil {
		return
	}

	if err = siteStream.novelStream(ctx); err != nil {
		return
	}

	if err = siteStream.chapterStream(ctx); err != nil {
		return
	}
	return
}

// 解析page url得到每一个novel的url
// from: catalogPage stream => novel stream
func (d DefaultSiteStreamImpl) catalogPageStream(ctx context.Context) error {
	source, err := NewRedisStreamSource(context.Background(), system.GetSystem().RedisClient,
		d.params.CatalogPageStreamName, d.params.CatalogPageStreamConsumer)
	if err != nil {
		return err
	}

	err = system.GetSystem().TaskPool.Submit(func() {
		source.
			Via(flow.NewMap(d.pr.HandleCatalogPageTask, uint(base.GetConfig().CrawlerSettings.CatalogPageTaskParallelism))).
			Via(flow.NewFlatMap(func(novelMsg []entity.NovelTask) []entity.NovelTask {
				return novelMsg
			}, 1)).
			To(NewRedisStreamSink(ctx, system.GetSystem().RedisClient,
				d.params.NovelPageStreamName))
	})
	if err != nil {
		zap.S().Error("failed to submit task", zap.Error(err))
		return err
	}
	return nil
}

// 处理每一个novel
func (d DefaultSiteStreamImpl) novelStream(ctx context.Context) error {
	source, err := NewRedisStreamSource(context.Background(), system.GetSystem().RedisClient,
		d.params.NovelPageStreamName, d.params.NovelPageStreamConsumer)
	if err != nil {
		return err
	}

	//item url
	sink := NewRedisStreamSink(ctx, system.GetSystem().RedisClient,
		d.params.ChapterPageStreamName)

	err = system.GetSystem().TaskPool.Submit(func() {
		source.
			Via(flow.NewMap(d.pr.HandleNovelTask, uint(base.GetConfig().CrawlerSettings.NovelTaskParallelism))).
			Via(flow.NewFlatMap(func(novelMsg []entity.ChapterTask) []entity.ChapterTask {
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
func (d DefaultSiteStreamImpl) chapterStream(ctx context.Context) error {
	source, err := NewRedisStreamSource(ctx, system.GetSystem().RedisClient,
		d.params.ChapterPageStreamName, d.params.ChapterPageStreamConsumer)
	if err != nil {
		return err
	}

	err = system.GetSystem().TaskPool.Submit(func() {
		source.
			Via(flow.NewMap(d.pr.HandleChapterTask, uint(base.GetConfig().CrawlerSettings.ChapterTaskParallelism))).
			To(extension.NewIgnoreSink())
	})
	if err != nil {
		zap.L().Error("failed to submit task", zap.Error(err))
		return err
	}
	return nil
}
