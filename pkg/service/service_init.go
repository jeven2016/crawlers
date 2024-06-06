package service

var ConfigService ConfigServiceInterface
var NovelService NovelServiceInterface
var SiteService SiteServiceInterface
var CatalogService CatalogServiceInterface
var CatalogPageTaskService CatalogPageTaskServiceInterface
var NovelTaskService NovelTaskServiceInterface
var ChapterService ChapterServiceInterface
var ChapterTaskService ChapterTaskServiceInterface
var ContentService ContentServiceInterface

func InitServices() {
	ConfigService = NewConfigService()
	NovelService = NewNovelService()
	SiteService = NewSiteService()
	CatalogService = NewCatalogService()
	CatalogPageTaskService = NewCatalogPageTaskService()
	NovelTaskService = NewNovelTaskService()
	ChapterService = NewChapterService()
	ChapterTaskService = NewChapterTaskService()
	ContentService = NewContentService()
}
