package repository

var CatalogDao catalogInterface
var SiteDao siteInterface
var CatalogPageTaskDao catalogPageTaskInterface
var NovelTaskDao novelTaskInterface
var NovelDao novelInterface
var ChapterDao chapterInterface
var ChapterTaskDao chapterTaskInterface
var ContentDao contentInterface

// InitRepositories initializes all the repository interfaces with their respective implementations.
// This function should be called once during the application startup to ensure all repositories are ready for use.
func InitRepositories() {
	// Initialize CatalogDao with catalogDaoImpl struct
	CatalogDao = &catalogDaoImpl{}

	// Initialize SiteDao with siteDaoImpl struct
	SiteDao = &siteDaoImpl{}

	// Initialize CatalogPageTaskDao with catalogPageTaskDaoImpl struct
	CatalogPageTaskDao = &catalogPageTaskDaoImpl{}

	// Initialize NovelTaskDao with novelTaskDaoImpl struct
	NovelTaskDao = &novelTaskDaoImpl{}

	// Initialize NovelDao with novelDaoImpl struct
	NovelDao = &novelDaoImpl{}

	// Initialize ChapterDao with chapterDaoImpl struct
	ChapterDao = &chapterDaoImpl{}

	// Initialize ChapterTaskDao with chapterTaskDaoImpl struct
	ChapterTaskDao = &chapterTaskDaoImpl{}

	// Initialize ContentDao with contentDaoImpl struct
	ContentDao = &contentDaoImpl{}
}
