package dao

import "context"

var CatalogDao catalogInterface
var SiteDao siteInterface
var CatalogPageTaskDao catalogPageTaskInterface
var NovelTaskDao novelTaskInterface
var NovelDao novelInterface
var ChapterDao chapterInterface
var ChapterTaskDao chapterTaskInterface
var ContentDao contentInterface

// InitDao initialize for db
func InitDao(ctx context.Context) {
	EnsureMongoIndexes(ctx)

	CatalogDao = &catalogDaoImpl{}
	SiteDao = &siteDaoImpl{}
	CatalogPageTaskDao = &catalogPageTaskDaoImpl{}
	NovelTaskDao = &novelTaskDaoImpl{}
	NovelDao = &novelDaoImpl{}
	ChapterDao = &chapterDaoImpl{}
	ChapterTaskDao = &chapterTaskDaoImpl{}
	ContentDao = &contentDaoImpl{}
}
