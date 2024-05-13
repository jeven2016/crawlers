package entity

import (
	"crawlers/pkg/base"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Site struct {
	Id          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name" binding:"required"`
	DisplayName string                 `bson:"displayName" json:"displayName" binding:"required"`
	Description string                 `bson:"description" json:"description"`
	Attributes  map[string]interface{} `bson:"attributes" json:"attributes"`
	CrawlerType base.CrawlerType       `bson:"crawlerType" json:"crawlerType" binding:"required"` //资源抓取类型

	CreatedTime *time.Time `bson:"created" bson:"createdTime"`
	UpdatedTime *time.Time `bson:"updated" bson:"updatedTime"`
}

type Catalog struct {
	Id          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	SiteId      primitive.ObjectID     `bson:"siteId,omitempty" json:"siteId" binding:"required"`
	Name        string                 `bson:"name" json:"name" binding:"required"`
	Description string                 `bson:"description" json:"description"`
	Attributes  map[string]interface{} `bson:"attributes" json:"attributes"`
	CrawlerType base.CrawlerType       `bson:"crawlerType" json:"crawlerType"` //资源抓取类型

	CreatedTime *time.Time `bson:"created" bson:"createdTime"`
	UpdatedTime *time.Time `bson:"updated" bson:"updatedTime"`
}

type Novel struct {
	Id          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	CatalogId   primitive.ObjectID     `bson:"catalogId,omitempty" json:"catalogId" binding:"required"`
	Name        string                 `bson:"name" json:"name" binding:"required"`
	Order       int                    `bson:"order" json:"order"`
	HasChapters bool                   `bson:"hasChapters" json:"hasChapters"`
	Description string                 `bson:"description" json:"description"`
	Attributes  map[string]interface{} `bson:"attributes" json:"attributes"`

	CreatedTime *time.Time `bson:"created" bson:"createdTime"`
	UpdatedTime *time.Time `bson:"updated" bson:"updatedTime"`
}

type Chapter struct {
	Id         primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	NovelId    primitive.ObjectID     `bson:"novelId,omitempty" json:"novelId" binding:"required"`
	Name       string                 `bson:"name" json:"name" binding:"required"`
	Order      int                    `bson:"order" json:"order"`
	Attributes map[string]interface{} `bson:"attributes" json:"attributes"`

	CreatedTime *time.Time `bson:"created" bson:"createdTime"`
	UpdatedTime *time.Time `bson:"updated" bson:"updatedTime"`
}

type Content struct {
	Id         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ParentId   primitive.ObjectID `bson:"parentId,omitempty" json:"parentId"`
	ParentType string             `bson:"parentType,omitempty" json:"parentType"` //chapter or novel
	Page       int                `bson:"page,omitempty" json:"page"`
	Content    string             `bson:"content" json:"content"`

	CreatedTime *time.Time `bson:"created" bson:"createdTime"`
	UpdatedTime *time.Time `bson:"updated" bson:"updatedTime"`
}

type RegexSettings struct {
	ParsePageRegex string `koanf:"parsePageRegex"`
	PagePrefix     string `koanf:"pagePrefix"`
	PageSuffix     string `koanf:"pageSuffix"`
}

type MongoCollections struct {
	Novel       string `koanf:"novel"`
	CatalogPage string `koanf:"catalogPage"`
}

type CrawlerSetting struct {
	Catalog     map[string]any `koanf:"catalog"`
	CatalogPage map[string]any `koanf:"catalogPage"`
	Novel       map[string]any `koanf:"novel"`
	Chapter     map[string]any `koanf:"chapter"`
}

type SiteSetting struct {
	Name             string            `koan:"name"`
	RegexSettings    *RegexSettings    `koanf:"regexSettings"`
	MongoCollections *MongoCollections `koanf:"mongoCollections"`
	Attributes       map[string]string `koanf:"attributes"`
	CrawlerSettings  *CrawlerSetting   `koanf:"crawlerSettings"`

	//whether to transfer redis message via separated redis streamuse separate space
	UseSeparateSpace bool `koanf:"useSeparateSpace"`
}

type CrawlerSettings struct {
	CatalogPageTaskParallelism int      `koanf:"catalogPageTaskParallelism"`
	NovelTaskParallelism       int      `koanf:"novelTaskParallelism"`
	ChapterTaskParallelism     int      `koanf:"chapterTaskParallelism"`
	ExcludedNovelUrls          []string `koanf:"excludedNovelUrls"`
}
