package stream

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/extension/sites/aipic"
	"crawlers/pkg/extension/sites/cartoon18"
	"crawlers/pkg/extension/sites/crawlers"
	nfs "crawlers/pkg/extension/sites/nsf"
	"crawlers/pkg/extension/sites/onej"
	"crawlers/pkg/model/entity"
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

func init() {
	siteCrawlerMap[base.Cartoon18] = cartoon18.NewCartoonCrawler()
	siteCrawlerMap[base.Kxkm] = crawlers.NewKxkmCrawler()
	siteCrawlerMap[base.Wucomic] = crawlers.NewWucomicCrawler()
	siteCrawlerMap[base.SiteNsf] = nfs.NewNsfCrawler()
	siteCrawlerMap[base.SiteOneJ] = onej.NewSiteOnej()
	siteCrawlerMap[base.Aipic] = aipic.NewCartoonCrawler()
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
