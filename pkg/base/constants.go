package base

import "errors"

const (
	RedisStreamDataVar = "data"
	SiteOneJ           = "onej"
	SiteYzs8           = "yzs8"
	SiteNsf            = "nsf"
	Cartoon18          = "cartoon18"
	Kxkm               = "kxkm"

	CollyMaxRetries = 3

	BulkBatchSize = 10 // 批量保存的文档数量

	ParentTypeChapter = "chapter"
	ParentTypeNovel   = "novel"
)

// db column
const (
	ColumId         = "_id"
	ColumnName      = "name"
	ColumnCatalogId = "catalogId"
	ColumnUrl       = "testUrl"
	ColumnSiteName  = "siteName"
	ColumnNovelId   = "novelId"
	ColumnParentId  = "parentId"
	ColumnPageNo    = "page"

	AttrAuthor = "author"
)

// db collection
const (
	CollectionSite            = "site"
	CollectionCatalog         = "catalog"
	CollectionNovel           = "novel"
	CollectionNovelTask       = "novelTask"
	CollectionChapter         = "chapter"
	CollectionChapterTask     = "chapterTask"
	CollectionCatalogPageTask = "catalogPageTask"
	CollectionContent         = "content"
)

var ConfigFiles = []string{"/etc/crawlers/crawlers.yaml"}

type CrawlerResourceType int

const (
	SiteResourceType CrawlerResourceType = iota + 1
	CatalogResourceType
	CatalogPageResourceType
	ArticleResourceType
	ArticlePageResourceType
	ChapterResourceType
	ChapterPageResourceType
)

// CrawlerType 抓取资源类型
type CrawlerType int

const (
	BtCrawlerType = iota + 1
	ComicCrawlerType
	NovelCrawlerType
)

// cache key prefix

const (
	SiteKeyExistsPrefix    = "site:exists"
	CatalogKeyExistsPrefix = "catalog:exists"
)

type TaskStatus int

const (
	TaskStatusNotStared TaskStatus = iota + 1
	TaskStatusProcessing
	TaskStatusFinished
	TaskStatusFailed
	TaskStatusRetryFailed
)

var ErrDecodingDocument = errors.New("document retrieved without decoding process")
var ErrDuplicatedDocument = errors.New("document is duplicated")
var ErrDocumentIdExists = errors.New("document's ID exists")

const DefaultRetries = 3
