package website

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/model"
	"crawlers/pkg/website/cartoon18"
	"crawlers/pkg/website/crawlers"
	nfs "crawlers/pkg/website/nsf"
	"crawlers/pkg/website/onej"
	"go.uber.org/zap"
)

type SiteCrawler interface {
	CrawlHomePage(ctx context.Context, url string) error
	CrawlCatalogPage(ctx context.Context, catalogPageMsg *model.CatalogPageTask) ([]model.NovelTask, error)
	CrawlNovelPage(ctx context.Context, novelPageMsg *model.NovelTask, skipSaveIfPresent bool) ([]model.ChapterTask, error)
	CrawlChapterPage(ctx context.Context, chapterMsg *model.ChapterTask, skipSaveIfPresent bool) error
}

// customized processors should be registered
var siteCrawlerMap = make(map[string]SiteCrawler)
var siteTaskProcessorMap = make(map[string]TaskProcessor)

func RegisterProcessors() {
	siteCrawlerMap[base.SiteOneJ] = onej.NewSiteOnej()
	siteCrawlerMap[base.SiteNsf] = nfs.NewNsfCrawler()
	siteCrawlerMap[base.Cartoon18] = cartoon18.NewCartoonCrawler()
	siteCrawlerMap[base.Kxkm] = crawlers.NewKxkmCrawler()

	//defaultTaskProcessor := NewTaskProcessor()
	//siteTaskProcessorMap[common.SiteOneJ] = defaultTaskProcessor
	//siteTaskProcessorMap[common.SiteNsf] = defaultTaskProcessor
	//siteTaskProcessorMap[common.Cartoon18] = defaultTaskProcessor
	//siteTaskProcessorMap[common.Kxkm] = defaultTaskProcessor
}

func GetSiteCrawler(siteName string) SiteCrawler {
	return siteCrawlerMap[siteName]
}

func GetSiteTaskProcessor(siteName string) TaskProcessor {
	if pr, ok := siteTaskProcessorMap[siteName]; ok {
		return pr
	}
	zap.L().Info("the default processor takes effect", zap.String("siteName", siteName))

	//return default processor
	return NewTaskProcessor()
}
