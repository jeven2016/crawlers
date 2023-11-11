package processor

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/model/entity"
	"crawlers/pkg/processor/sites/cartoon18"
	"crawlers/pkg/processor/sites/crawlers"
	"crawlers/pkg/processor/sites/nsf"
	"crawlers/pkg/processor/sites/onej"
	"go.uber.org/zap"
)

type SiteCrawler interface {
	CrawlHomePage(ctx context.Context, url string) error
	CrawlCatalogPage(ctx context.Context, catalogPageMsg *entity.CatalogPageTask) ([]entity.NovelTask, error)
	CrawlNovelPage(ctx context.Context, novelPageMsg *entity.NovelTask, skipSaveIfPresent bool) ([]entity.ChapterTask, error)
	CrawlChapterPage(ctx context.Context, chapterMsg *entity.ChapterTask, skipSaveIfPresent bool) error
}

// customized processors should be registered
var siteCrawlerMap = make(map[string]SiteCrawler)
var siteTaskProcessorMap = make(map[string]TaskProcessor)

func RegisterProcessors() {
	siteCrawlerMap[base.SiteOneJ] = onej.NewSiteOnej()
	siteCrawlerMap[base.SiteNsf] = nfs.NewNsfCrawler()
	siteCrawlerMap[base.Cartoon18] = cartoon18.NewCartoonCrawler()
	siteCrawlerMap[base.Kxkm] = crawlers.NewKxkmCrawler()
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