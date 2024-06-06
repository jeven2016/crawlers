package repository

var CatalogRepo catalogRepo
var SiteRepo siteRepo
var CatalogPageTaskRepo catalogPageTaskRepo
var NovelTaskRepo novelTaskRepo
var NovelRepo novelRepo
var ChapterRepo chapterRepo
var ChapterTaskRepo chapterTaskRepo
var ContentRepo contentRepo

// InitRepositories initializes all the repository interfaces with their respective implementations.
// This function should be called once during the application startup to ensure all repositories are ready for use.
func InitRepositories() {
	// Initialize CatalogRepo with catalogRepoImpl struct
	CatalogRepo = &catalogRepoImpl{}

	// Initialize SiteRepo with siteRepoImpl struct
	SiteRepo = &siteRepoImpl{}

	// Initialize CatalogPageTaskRepo with catalogPageTaskRepoImpl struct
	CatalogPageTaskRepo = &catalogPageTaskRepoImpl{}

	// Initialize NovelTaskRepo with novelTaskRepoImpl struct
	NovelTaskRepo = &novelTaskRepoImpl{}

	// Initialize NovelRepo with novelRepoImpl struct
	NovelRepo = &novelRepoImpl{}

	// Initialize ChapterRepo with chapterRepoImpl struct
	ChapterRepo = &chapterRepoImpl{}

	// Initialize ChapterTaskRepo with chapterTaskRepoImpl struct
	ChapterTaskRepo = &chapterTaskRepoImpl{}

	// Initialize ContentRepo with contentRepoImpl struct
	ContentRepo = &contentRepoImpl{}
}
