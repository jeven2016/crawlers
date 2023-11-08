package dto

type CatalogPageRequest struct {
	SiteKey string `json:"siteKey"`
	Catalog string `json:"catalog"`
	PageUrl string `json:"pageUrl"`
}

type CreateRequest struct {
	Key           string `json:"key"`
	Name          string `json:"name"`
	Entity        any    `json:"entity"`
	Collection    string `json:"collection"`
	RedisCacheKey string `json:"redisCacheKey"`
}
