package base

import "github.com/jeven2016/mylibs/config"

type Config interface {
	GetServerConfig() *ServerConfig
	Validate() error
	Complete() error
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

type SiteConfig struct {
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

type ServerConfig struct {
	config.ServerConfig `koanf:",squash"`
	CrawlerSettings     *CrawlerSettings `koanf:"crawlerSettings"`
	WebSites            []SiteConfig     `koanf:"webSites"`
}

func (s ServerConfig) GetServerConfig() *config.ServerConfig {
	return &s.ServerConfig
}
func (s ServerConfig) Validate() error {
	return nil
}
func (s ServerConfig) Complete() error {
	return nil
}
