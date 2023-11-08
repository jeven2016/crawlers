package model

import (
	"crawlers/pkg/base"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Resource 资源，标记网站上需要下载的的资源，供爬虫使用
type Resource interface {
	ResourceType() base.CrawlerResourceType
	GetUrl() string
	GetStatus() base.TaskStatus
}

// OperationDate 操作日期
type OperationDate struct {
	CreatedDate *time.Time `bson:"createdDate" json:"createdDate"`
	LastUpdated *time.Time `bson:"lastUpdated" json:"lastUpdated"`
}

// SiteTask 站点
type SiteTask struct {
	Id         primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name       string                 `bson:"name" json:"name" binding:"required"`
	Url        string                 `bson:"url" json:"url" binding:"required"`
	Attributes map[string]interface{} `bson:"attributes" json:"attributes"`
	Status     base.TaskStatus        `bson:"status" json:"status"`
	Retries    uint32                 `bson:"retries" json:"retries"`

	OperationDate
}

func (s *SiteTask) ResourceType() base.CrawlerResourceType {
	return base.SiteResourceType
}

func (s *SiteTask) GetUrl() string {
	return s.Url
}

func (s *SiteTask) GetStatus() base.TaskStatus {
	return s.Status
}

// CatalogTask 分类目录
type CatalogTask struct {
	// 添加omitempty，当为空时，mongo driver会自动生成
	Id              primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	ParentCatalogId primitive.ObjectID     `bson:"parentCatalogId,omitempty" json:"parentCatalogId"`
	Name            string                 `bson:"name" json:"name" binding:"required"`
	Url             string                 `bson:"url" json:"url" binding:"required"`
	Attributes      map[string]interface{} `bson:"attributes" json:"attributes"`
	Status          base.TaskStatus        `bson:"status" json:"status"`
	Retries         uint32                 `bson:"retries" json:"retries"`
	SiteName        string                 `bson:"siteName" json:"siteName"` //便于后续的日志输出
	OperationDate
}

func (s *CatalogTask) ResourceType() base.CrawlerResourceType {
	return base.CatalogResourceType
}
func (s *CatalogTask) GetUrl() string {
	return s.Url
}
func (s *CatalogTask) GetStatus() base.TaskStatus {
	return s.Status
}

type CatalogPageTask struct {
	// 添加omitempty，当为空时，mongo driver会自动生成
	Id         primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	CatalogId  primitive.ObjectID     `json:"catalogId" bson:"catalogId" binding:"required"`
	Url        string                 `bson:"url" json:"url" binding:"required" binding:"required"`
	Attributes map[string]interface{} `bson:"attributes" json:"attributes"`
	Status     base.TaskStatus        `bson:"status" json:"status"`
	SiteName   string                 `bson:"siteName" json:"siteName"`
	Retries    uint32                 `bson:"retries" json:"retries"`
	OperationDate
}

func (c CatalogPageTask) ResourceType() base.CrawlerResourceType {
	return base.CatalogPageResourceType
}
func (c CatalogPageTask) GetUrl() string {
	return c.Url
}
func (c CatalogPageTask) GetStatus() base.TaskStatus {
	return c.Status
}

type NovelTask struct {
	Id          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name"`
	CatalogId   primitive.ObjectID     `bson:"catalogId,omitempty" json:"catalogId" binding:"required"`
	Url         string                 `bson:"url" json:"url" binding:"required"`
	HasChapters bool                   `bson:"hasChapters,omitempty" json:"hasChapters"`
	Attributes  map[string]interface{} `bson:"attributes" json:"attributes"`
	Status      base.TaskStatus        `bson:"status" json:"status"`
	Retries     uint32                 `bson:"retries" json:"retries"`
	SiteName    string                 `bson:"siteName" json:"siteName"`
	OperationDate
}

func (s *NovelTask) ResourceType() base.CrawlerResourceType {
	return base.ArticleResourceType
}
func (s *NovelTask) GetUrl() string {
	return s.Url
}
func (s *NovelTask) GetStatus() base.TaskStatus {
	return s.Status
}

type NovelPageTask struct {
	// 添加omitempty，当为空时，mongo driver会自动生成
	Id         primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	ArticleId  primitive.ObjectID     `bson:"articleId,omitempty" json:"articleId" binding:"required"`
	Url        string                 `bson:"url,omitempty" json:"url" binding:"required"`
	Attributes map[string]interface{} `bson:"attributes" json:"attributes"`
	Status     base.TaskStatus        `bson:"status" json:"status"`
	Retries    uint32                 `bson:"retries" json:"retries"`
	SiteName   string                 `bson:"siteName" json:"siteName"`
	OperationDate
}

func (s *NovelPageTask) ResourceType() base.CrawlerResourceType {
	return base.ArticlePageResourceType
}
func (s *NovelPageTask) GetUrl() string {
	return s.Url
}
func (s *NovelPageTask) GetStatus() base.TaskStatus {
	return s.Status
}

type ChapterTask struct {
	Id       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name"`
	Order    int                `bson:"order" json:"order"`
	NovelId  primitive.ObjectID `bson:"novelId,omitempty" json:"novelId"`
	Url      string             `bson:"url" json:"url"`
	Status   base.TaskStatus    `bson:"status" json:"status"`
	Retries  uint32             `bson:"retries" json:"retries"`
	SiteName string             `bson:"siteName" json:"siteName"`
	OperationDate
}

func (s *ChapterTask) ResourceType() base.CrawlerResourceType {
	return base.ChapterResourceType
}
func (s *ChapterTask) GetUrl() string {
	return s.Url
}
func (s *ChapterTask) GetStatus() base.TaskStatus {
	return s.Status
}

type ChapterPageTask struct {
	// 添加omitempty，当为空时，mongo driver会自动生成
	Id         primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	ArticleId  primitive.ObjectID     `bson:"articleId,omitempty" json:"articleId"`
	Url        string                 `bson:"url,omitempty" json:"url"`
	Attributes map[string]interface{} `bson:"attributes" json:"attributes"`
	Status     base.TaskStatus        `bson:"status" json:"status"`
	Retries    uint32                 `bson:"retries" json:"retries"`
	SiteName   string                 `bson:"siteName" json:"siteName"`
	OperationDate
}

func (s *ChapterPageTask) ResourceType() base.CrawlerResourceType {
	return base.ChapterPageResourceType
}
func (s *ChapterPageTask) GetUrl() string {
	return s.Url
}
func (s *ChapterPageTask) GetStatus() base.TaskStatus {
	return s.Status
}
