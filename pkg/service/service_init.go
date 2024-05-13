package service

var ConfigService ConfigServiceInterface
var NovelService NovelServiceInterface

func InitServices() {
	ConfigService = NewConfigService()
	NovelService = NewNovelService()

}
