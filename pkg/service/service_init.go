package service

var ConfigService ConfigServiceInterface
var NovelService NovelServiceInterface
var SiteService SiteServiceInterface

func InitServices() {
	ConfigService = NewConfigService()
	NovelService = NewNovelService()
	SiteService = NewSiteService()
}
