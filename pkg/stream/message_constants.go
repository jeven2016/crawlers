package stream

// site:
//
//	  Catalog
//	      CatalogPage:
//			   Item:
//				  Chapter:
//					  body

const (
	HomeUrlStream         = "homeUrlStream"
	HomeUrlStreamConsumer = "HomeUrlStreamConsumer"

	HomeSiteKey = "homeSiteKey"

	SiteCatalogHomeUrlStream         = "SiteCatalogHomeUrlStream"
	SiteCatalogHomeUrlStreamConsumer = "SiteCatalogHomeUrlStreamConsumer"

	// 某个catalog下的某一页
	CatalogPageUrlStream         = "CatalogPageUrlStream"
	CatalogPageUrlStreamConsumer = "CatalogPageUrlStreamConsumer"

	//每页中解析出的具体item
	NovelUrlStream         = "NovelUrlStream"
	NovelUrlStreamConsumer = "NovelUrlStreamConsumer"

	ChapterUrlStream         = "ChapterUrlStream"
	ChapterUrlStreamConsumer = "ChapterUrlStreamConsumer"
)
