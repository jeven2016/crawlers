package dao

var CatalogDao catalogInterface
var SiteDao siteInterface
var CatalogPageTaskDao catalogPageTaskInterface
var NovelTaskDao novelTaskInterface
var NovelDao novelInterface
var ChapterDao chapterInterface
var ChapterTaskDao chapterTaskInterface
var ContentDao contentInterface

// init initialize for db
func init() {
	CatalogDao = &catalogDaoImpl{}
	SiteDao = &siteDaoImpl{}
	CatalogPageTaskDao = &catalogPageTaskDaoImpl{}
	NovelTaskDao = &novelTaskDaoImpl{}
	NovelDao = &novelDaoImpl{}
	ChapterDao = &chapterDaoImpl{}
	ChapterTaskDao = &chapterTaskDaoImpl{}
	ContentDao = &contentDaoImpl{}
}
