package api

import (
	_ "crawlers/docs"
	"crawlers/pkg/api/handler"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// RegisterEndpoints register all web endpoints
func RegisterEndpoints(localeCfg gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	var engine = gin.Default()

	// apply i18n middleware
	engine.Use(localeCfg)

	// Add a ginzap middleware, which:
	//   - Logs all requests, like a combined access and error log.
	//   - Logs to stdout.
	//   - RFC3339 with local time format.
	engine.Use(ginzap.Ginzap(zap.L(), time.RFC3339, false))

	// Logs all panic to error log
	//   - stack means whether output the stack info.
	engine.Use(ginzap.RecoveryWithZap(zap.L(), false))

	hd := handler.NewTaskHandler()
	siteHandler := handler.NewSiteHandler()

	//gin-swagger 同时还提供了 DisablingWrapHandler 函数，方便我们通过设置某些环境变量来禁用Swagger。
	//此时如果将环境变量 NAME_OF_ENV_VARIABLE设置为任意值，则 /swagger/*any 将返回404响应，就像未指定路由时一样
	//engine.GET("/swagger/*any", ginSwagger.DisablingWrapHandler(swaggerFiles.taskHandler, "NAME_OF_ENV_VARIABLE"))
	// no prefix /api/v1
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	engine.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// with prefix /api/v1
	routerGroup := engine.Group("/api/v1")

	routerGroup.GET("/sites", siteHandler.FindSites)
	routerGroup.GET("/sites/:siteId", siteHandler.FindSiteById)
	routerGroup.DELETE("/sites/:siteId", siteHandler.DeleteSite)
	routerGroup.GET("/sites/:siteId/catalogs", siteHandler.FindSiteCatalogs)

	routerGroup.GET("/sites/:siteId/settings", siteHandler.FindSiteSettings)

	routerGroup.GET("/tasks/catalog-pages", hd.FindTasksOfCatalogPage)
	routerGroup.GET("/tasks/novels", hd.FindTasksOfNovel)

	routerGroup.POST("/catalogs", siteHandler.CreateCatalog)
	routerGroup.POST("/catalogs/:catalogId", siteHandler.FindCatalogById)
	routerGroup.POST("/sites", siteHandler.CreateSite)
	routerGroup.POST("/tasks/catalog-pages", hd.CreateCatalogPageTask)
	routerGroup.POST("/tasks/novels", hd.CreateNovelPageTask)

	routerGroup.DELETE("/tasks/novels", hd.DeleteNovelPageTasks)
	//routerGroup.POST("/tasks/schedule-task", hd.RunScheduleTask)

	return engine
}
