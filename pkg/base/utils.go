package base

import (
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"os"
)

// PrintCmdErr print error in console
func PrintCmdErr(err error) {
	_, err = fmt.Fprintf(os.Stderr, "Error: '%s' \n", err)
	if err != nil {
		panic(err)
	}
}

func GetSiteConfig(siteKey string) *SiteConfig {
	cfg, ok := slice.FindBy(GetConfig().WebSites, func(index int, item SiteConfig) bool {
		return item.Name == siteKey
	})
	if !ok {
		return nil
	}
	return &cfg
}
