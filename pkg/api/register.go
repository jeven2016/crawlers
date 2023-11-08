package api

import (
	_ "crawlers/docs"
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
func RegisterEndpoints() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	var engine = gin.Default()

	logger := zap.L()

	// Add a ginzap middleware, which:
	//   - Logs all requests, like a combined access and error log.
	//   - Logs to stdout.
	//   - RFC3339 with local time format.
	engine.Use(ginzap.Ginzap(logger, time.RFC3339, false))

	// Logs all panic to error log
	//   - stack means whether output the stack info.
	engine.Use(ginzap.RecoveryWithZap(logger, false))

	hd := NewTaskHandler()
	siteHandler := NewSiteHandler()

	//gin-swagger 同时还提供了 DisablingWrapHandler 函数，方便我们通过设置某些环境变量来禁用Swagger。
	//此时如果将环境变量 NAME_OF_ENV_VARIABLE设置为任意值，则 /swagger/*any 将返回404响应，就像未指定路由时一样
	//engine.GET("/swagger/*any", ginSwagger.DisablingWrapHandler(swaggerFiles.TaskHandler, "NAME_OF_ENV_VARIABLE"))
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	engine.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	engine.POST("/catalogs", siteHandler.CreateCatalog)
	engine.POST("/sites", siteHandler.CreateSite)
	engine.POST("/tasks/catalog-pages", hd.HandleCatalogPage)
	engine.POST("/tasks/novels", hd.HandleNovelPage)
	engine.POST("/tasks/schedule-task", hd.RunScheduleTask)

	return engine
}
