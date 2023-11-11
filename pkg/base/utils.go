package base

import (
	"github.com/duke-git/lancet/v2/slice"
)

func GetSiteConfig(siteKey string) *SiteConfig {
	cfg, ok := slice.FindBy(GetConfig().WebSites, func(index int, item SiteConfig) bool {
		return item.Name == siteKey
	})
	if !ok {
		return nil
	}
	return &cfg
}
