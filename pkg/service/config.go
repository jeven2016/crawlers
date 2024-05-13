package service

import (
	"crawlers/pkg/model/entity"
	"github.com/jeven2016/mylibs/config"
)

type InternalConfig struct {
	config.ServerConfig `koanf:",squash"`
	CrawlerSettings     *entity.CrawlerSettings `koanf:"crawlerSettings"`
	WebSites            []entity.SiteSetting    `koanf:"webSites"`
}
