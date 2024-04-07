package api

import (
	_ "crawlers/docs"
	"embed"
	"encoding/json"
	ginI18n "github.com/gin-contrib/i18n"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"golang.org/x/text/language"
	"net/http"
	"time"
)

// RegisterEndpoints register all web endpoints
func RegisterEndpoints(i18nFs embed.FS) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	var engine = gin.Default()

	// apply i18n middleware
	engine.Use(ginI18n.Localize(ginI18n.WithBundle(&ginI18n.BundleCfg{
		DefaultLanguage:  language.Chinese,
		FormatBundleFile: "json",
		AcceptLanguage:   []language.Tag{language.Chinese},
		RootPath:         "./pkg/i18n/",
		UnmarshalFunc:    json.Unmarshal,

		//get resource from embedded bundle file
		Loader: &ginI18n.EmbedLoader{
			FS: i18nFs,
		},
	})))

	// Add a ginzap middleware, which:
	//   - Logs all requests, like a combined access and error log.
	//   - Logs to stdout.
	//   - RFC3339 with local time format.
	engine.Use(ginzap.Ginzap(zap.L(), time.RFC3339, false))

	// Logs all panic to error log
	//   - stack means whether output the stack info.
	engine.Use(ginzap.RecoveryWithZap(zap.L(), false))

	hd := NewTaskHandler()
	siteHandler := NewSiteHandler()

	//gin-swagger 同时还提供了 DisablingWrapHandler 函数，方便我们通过设置某些环境变量来禁用Swagger。
	//此时如果将环境变量 NAME_OF_ENV_VARIABLE设置为任意值，则 /swagger/*any 将返回404响应，就像未指定路由时一样
	//engine.GET("/swagger/*any", ginSwagger.DisablingWrapHandler(swaggerFiles.taskHandler, "NAME_OF_ENV_VARIABLE"))
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	engine.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	engine.GET("/sites", siteHandler.FindSites)
	engine.GET("/sites/:siteId/catalogs", siteHandler.GetSiteCatalogs)
	engine.GET("/tasks/catalog-pages", hd.GetTasksOfCatalogPage)
	engine.GET("/tasks/novels", hd.GetTasksOfNovel)

	engine.POST("/catalogs", siteHandler.CreateCatalog)
	engine.POST("/sites", siteHandler.CreateSite)
	engine.POST("/tasks/catalog-pages", hd.HandleCatalogPage)
	engine.POST("/tasks/novels", hd.HandleNovelPage)
	//engine.POST("/tasks/schedule-task", hd.RunScheduleTask)

	return engine
}
