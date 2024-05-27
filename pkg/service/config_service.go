package service

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/model/entity"
	"errors"
	"github.com/duke-git/lancet/v2/slice"
	gconfig "github.com/jeven2016/mylibs/config"
	"github.com/jeven2016/mylibs/system"
	"github.com/jeven2016/mylibs/utils"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
)

type ConfigServiceInterface interface {
	GetConfig() *InternalConfig
	LoadInternalConfig(yamlConfig string, extraConfigFile *string) error
	MergeConfig()
	GetSiteConfig(siteName string) *entity.SiteSettings
}

type configServiceImpl struct {
	internalCfg   *InternalConfig
	siteConfigMap map[string]*entity.SiteSettings
}

func NewConfigService() ConfigServiceInterface {
	return &configServiceImpl{
		internalCfg:   nil,
		siteConfigMap: map[string]*entity.SiteSettings{},
	}
}

// GetSiteConfig retrieves the site configuration for the given siteKey.
// If the configuration is not found in the internal configuration, it will attempt to retrieve it from Redis.
// If the configuration is not found in Redis, it will create a default configuration, store it in Redis, and return it.
//
// Parameters:
// siteKey (string): The unique identifier for the site.
//
// Returns:
// *entity.SiteSettings: A pointer to the site configuration. If an error occurs during retrieval or creation, it returns nil.
func (c *configServiceImpl) GetSiteConfig(siteName string) *entity.SiteSettings {
	// Check if the site configuration is present in the internal configuration
	cfg, ok := slice.FindBy(c.internalCfg.WebSites, func(index int, item entity.SiteSettings) bool {
		return item.Name == siteName
	})
	if ok {
		return &cfg
	}

	// Generate the Redis key for the site configuration
	siteConfigKey := utils.GenKey("siteConfig", siteName)

	// Attempt to retrieve the site configuration from Redis
	value, err := system.GetSystem().RedisClient.Client.Get(context.Background(), siteConfigKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// If the configuration is not found in Redis, create a default configuration
			defaultSetting := c.createDefaultSiteSettings(siteName)

			// Marshal the default configuration to a byte array
			b, err := msgpack.Marshal(defaultSetting)
			if err != nil {
				zap.L().Warn("failed to marshal", zap.Error(err))
				return nil
			}

			// Store the default configuration in Redis
			if err = system.GetSystem().RedisClient.Client.Set(context.Background(), siteConfigKey, b, 0).Err(); err != nil {
				zap.L().Warn("failed to set", zap.Error(err))
				return nil
			}

			// Return the default configuration
			return defaultSetting
		}
		return nil
	} else {
		// If the configuration is found in Redis, unmarshal it from the byte array
		defaultSetting := &entity.SiteSettings{}
		if err = msgpack.Unmarshal([]byte(value), defaultSetting); err != nil {
			zap.L().Warn("failed to unmarshal", zap.Error(err))
			return nil
		}

		// Return the unmarshalled configuration
		return defaultSetting
	}
}

func (c *configServiceImpl) createDefaultSiteSettings(siteKey string) *entity.SiteSettings {
	defaultSetting := &entity.SiteSettings{
		Name:          siteKey + "_default",
		RegexSettings: &entity.RegexSettings{},
		MongoCollections: &entity.MongoCollections{
			Novel:       "novel_default",
			CatalogPage: "catalogPage_default",
		},
		Attributes: map[string]string{},
		CrawlerSettings: &entity.CrawlerSetting{
			Catalog: map[string]any{
				"skipIfPresent":     false,
				"skipSaveIfPresent": true,
			},
			CatalogPage: map[string]any{
				"skipIfPresent":     false,
				"skipSaveIfPresent": true,
			},
			Novel: map[string]any{
				"skipIfPresent":     false,
				"skipSaveIfPresent": true,
			},

			Chapter: map[string]any{
				"skipIfPresent":     false,
				"skipSaveIfPresent": true,
			},
		},
	}
	return defaultSetting
}

func (c *configServiceImpl) GetConfig() *InternalConfig {
	return c.internalCfg
}

func (c *configServiceImpl) LoadInternalConfig(yamlConfig string, extraConfigFile *string) error {
	//load internal config
	cfg := &InternalConfig{}
	if err := gconfig.LoadConfig([]byte(yamlConfig), cfg, extraConfigFile, base.ConfigFiles); err != nil {
		return err
	}
	c.internalCfg = cfg
	return nil
}

func (c *configServiceImpl) MergeConfig() {
	//TODO implement me
	panic("implement me")
}
