package stream

import (
	"crawlers/pkg/base"
	"go.uber.org/zap"
	"strings"
)

type StreamTaskParams struct {
	CatalogPageStreamName     string
	CatalogPageStreamConsumer string
	NovelPageStreamName       string
	NovelPageStreamConsumer   string
	ChapterPageStreamName     string
	ChapterPageStreamConsumer string
}

func DefaultStreamTaskParams() *StreamTaskParams {
	return &StreamTaskParams{
		CatalogPageStreamName:     CatalogPageUrlStream,
		CatalogPageStreamConsumer: CatalogPageUrlStreamConsumer,
		NovelPageStreamName:       NovelUrlStream,
		NovelPageStreamConsumer:   NovelUrlStreamConsumer,
		ChapterPageStreamName:     ChapterUrlStream,
		ChapterPageStreamConsumer: ChapterUrlStreamConsumer,
	}
}

func GenStreamTaskParams(siteName string) *StreamTaskParams {
	defaultParams := DefaultStreamTaskParams()
	if siteName == "" {
		return defaultParams
	}

	cfg := base.GetSiteConfig(siteName)
	if cfg == nil {
		zap.L().Warn("The default stream task parameters returned while no customized definition found",
			zap.String("siteName", siteName))
	} else if cfg.UseSeparateSpace {
		siteName = strings.TrimSpace(siteName)

		//of := reflect.TypeOf(defaultParams)
		//reflect.VisibleFields(of)
		defaultParams.CatalogPageStreamName = CatalogPageUrlStream + "_" + siteName
		defaultParams.CatalogPageStreamConsumer = CatalogPageUrlStreamConsumer + "_" + siteName
		defaultParams.NovelPageStreamName = NovelUrlStream + "_" + siteName
		defaultParams.NovelPageStreamConsumer = NovelUrlStreamConsumer + "_" + siteName
		defaultParams.ChapterPageStreamName = ChapterUrlStream + "_" + siteName
		defaultParams.ChapterPageStreamConsumer = ChapterUrlStreamConsumer + "_" + siteName
	}
	return defaultParams
}
