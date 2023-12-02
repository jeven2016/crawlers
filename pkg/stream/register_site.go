package stream

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/stream/processor"
	"github.com/jeven2016/mylibs/system"
	"github.com/reugn/go-streams"
	"github.com/reugn/go-streams/extension"
	"github.com/reugn/go-streams/flow"
	"go.uber.org/zap"
	"sync"
)

var siteStreamMap = map[string]SiteStreamInterface{}
var streamInitLock = sync.Mutex{}

// DefaultSiteStreamImpl Default site stream implementation
type DefaultSiteStreamImpl struct {
	pr     processor.TaskProcessor
	params *StreamTaskParams
}

// LaunchGlobalSiteStream launches site streams for global sharing
func LaunchGlobalSiteStream(ctx context.Context) error {
	return LaunchSiteStream(ctx, "")
}

// LaunchSiteStream launch separated streams for a site
func LaunchSiteStream(ctx context.Context, siteName string) error {
	if siteStreamMap[siteName] == nil {
		//initialize for this site
		streamInitLock.Lock()
		defer streamInitLock.Unlock()

		if siteStreamMap[siteName] == nil {
			siteStream := &DefaultSiteStreamImpl{
				params: GenStreamTaskParams(siteName),
				pr:     processor.GetSiteTaskProcessor(siteName),
			}

			funcSlice := []func(ctx2 context.Context) error{
				siteStream.catalogPageStream,
				siteStream.novelStream,
				siteStream.chapterStream,
			}

			for i := 0; i < len(funcSlice); i++ {
				if err := funcSlice[i](ctx); err != nil {
					return err
				}
			}
			zap.S().Infof("some background tasks are launched for the site " + siteName)
		}
	}
	return nil
}

// 解析page url得到每一个novel的url
// from: catalogPage stream => novel stream
func (d DefaultSiteStreamImpl) catalogPageStream(ctx context.Context) error {
	var sourceParallelism = base.GetConfig().CrawlerSettings.CatalogPageTaskParallelism
	var sinkParallelism = base.GetConfig().CrawlerSettings.NovelTaskParallelism
	flowFunction := flow.NewFlatMap(d.pr.HandleCatalogPageTask, uint(sourceParallelism))
	return createStream(ctx, d.params.CatalogPageStreamName, d.params.CatalogPageStreamConsumer,
		d.params.NovelPageStreamName, flowFunction, sourceParallelism, sinkParallelism, false)
}

// 处理每一个novel
func (d DefaultSiteStreamImpl) novelStream(ctx context.Context) error {
	var sourceParallelism = base.GetConfig().CrawlerSettings.NovelTaskParallelism
	var sinkParallelism = base.GetConfig().CrawlerSettings.ChapterTaskParallelism
	flowFunction := flow.NewFlatMap(d.pr.HandleNovelTask, uint(sourceParallelism))
	return createStream(ctx, d.params.NovelPageStreamName, d.params.NovelPageStreamConsumer,
		d.params.ChapterPageStreamName, flowFunction, sourceParallelism, sinkParallelism, false)
}

// 处理每一个novel
func (d DefaultSiteStreamImpl) chapterStream(ctx context.Context) error {
	var sourceParallelism = base.GetConfig().CrawlerSettings.ChapterTaskParallelism
	flowFunction := flow.NewMap(d.pr.HandleChapterTask, uint(sourceParallelism))
	return createStream(ctx, d.params.ChapterPageStreamName, d.params.ChapterPageStreamConsumer,
		d.params.ChapterPageStreamName, flowFunction, sourceParallelism, sourceParallelism, true)
}

// createStream creates specified stream
func createStream(
	ctx context.Context,
	sourceChanel string,
	consumerGroup string,
	sinkChanel string,
	flatMapFlow streams.Flow,
	sourceChanCapacity,
	sinkChanCapacity int,
	ignoredSink bool) error {
	source, err := NewRedisStreamSource(ctx, system.GetSystem().RedisClient, sourceChanel, consumerGroup, sourceChanCapacity)
	if err != nil {
		return err
	}

	err = system.GetSystem().TaskPool.Submit(func() {
		streamFlow := source.Via(flatMapFlow)

		if ignoredSink {
			streamFlow.To(extension.NewIgnoreSink())
		} else {
			sink := NewRedisStreamSink(ctx, system.GetSystem().RedisClient, sinkChanel, sinkChanCapacity)
			streamFlow.To(sink)
		}
	})
	if err != nil {
		return err
	}
	return nil
}
