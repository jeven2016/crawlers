package stream

import (
	"context"
	"github.com/reugn/go-streams/flow"
)

type SiteStreamInterface interface {
	catalogPageStream(ctx context.Context)
	novelStream(ctx context.Context)
	chapterStream(ctx context.Context)
}

type StreamStepDefinition[T, R, E, U any] struct {
	sourceStream        string
	sourceConsumerGroup string
	destinationStream   string
	convertFunc         flow.MapFunction[T, R]
	flowFlatMap         flow.FlatMap[E, U]
}

type StreamTaskParams struct {
	CatalogPageStreamName     string
	CatalogPageStreamConsumer string
	NovelPageStreamName       string
	NovelPageStreamConsumer   string
	ChapterPageStreamName     string
	ChapterPageStreamConsumer string
}
