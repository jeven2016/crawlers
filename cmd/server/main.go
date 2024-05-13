package main

import (
	"context"
	"crawlers/pkg/api"
	"crawlers/pkg/base"
	"crawlers/pkg/repository"
	"crawlers/pkg/service"
	"crawlers/pkg/stream"
	"embed"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	"github.com/jeven2016/mylibs/system"
	"github.com/jeven2016/mylibs/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/text/language"
	"net/http"
)

//go:embed internal_conf.yaml
var configFile string

//go:embed i18n/*
var i18nFs embed.FS

const softwareVersion = "0.1"
const flagName = "config"

var extraConfigFile *string

// @title  crawler文档
// @version 0.2
// @description  crawler接口参考文档
// @termsOfService only for internal use
// @BasePath /api/v1/
// @query.collection.format multi
func main() {
	//printBanner()
	run()
}

func run() {
	var server *http.Server
	var rootCmd = &cobra.Command{
		Version: softwareVersion,
		Use:     "crawlers",
		Short:   "crawlers",
		Run: func(cmd *cobra.Command, args []string) {
			runServer(server)
		},
	}

	// the absolute path of yaml config file
	extraConfigFile = rootCmd.Flags().StringP(flagName, "c", "", "the absolute path of yaml config file")

	if err := rootCmd.Execute(); err != nil {
		utils.PrintCmdErr(err)
	}
}

// runServer initializes and starts the web server.
// It loads the internal configuration, sets up the server, and initializes the global context.
// It also ensures the creation of MongoDB indexes, launches global streams, and handles any errors during server startup.
//
// Parameters:
// - server: A pointer to the http.Server struct representing the existing server.
func runServer(server *http.Server) {
	repository.InitRepositories()
	service.InitServices()

	//load internal config
	err := service.ConfigService.LoadInternalConfig(configFile, extraConfigFile)
	if err != nil {
		utils.PrintCmdErr(err)
		return
	}

	// globally cache the config
	server = createServer(server)

	//global context
	ctx, cancelFunc := context.WithCancel(context.Background())
	base.SetSystemContext(ctx)

	sys := systemInit(cancelFunc, server, ctx)
	if sys != nil {
		//ensure the indexes are created
		repository.EnsureMongoIndexes(ctx)

		//global streams
		if err := stream.LaunchGlobalSiteStream(ctx); err != nil {
			zap.L().Error("failed to register streams", zap.Error(err))
			system.Stop(ctx)
			return
		}

		// run a web server
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Error("unable to start web server", zap.Error(err))
		}
	}
}

// createServer initializes and configures a new HTTP server with Gin framework.
// It also sets up internationalization (i18n) using gin-contrib/i18n package.
//
// Parameters:
// - cfg: A pointer to the InternalConfig struct containing the server configuration.
// - server: A pointer to the http.Server struct representing the existing server.
//
// Returns:
// - A pointer to the configured http.Server struct.
func createServer(server *http.Server) *http.Server {
	// Initialize the i18n configuration
	localeCfg := ginI18n.Localize(ginI18n.WithBundle(
		&ginI18n.BundleCfg{
			DefaultLanguage:  language.Chinese,
			FormatBundleFile: "json",
			AcceptLanguage:   []language.Tag{language.Chinese},
			RootPath:         "./i18n/",
			UnmarshalFunc:    json.Unmarshal,

			// Load resource from embedded bundle file
			Loader: &ginI18n.EmbedLoader{
				FS: i18nFs,
			},
		}),
		ginI18n.WithGetLngHandle(
			// Set default language to Chinese
			// http://path/
			//
			// Set language to English
			// http://path/?lang=en
			//
			// Set language to Chinese
			// http://localhost:9014/?lang=zh
			func(context *gin.Context, defaultLng string) string {
				// Get language from query string
				lng := context.Query("lang")
				if lng == "" {
					return defaultLng
				}
				return lng
			},
		))

	// Register endpoints and apply i18n middleware
	engine := api.RegisterEndpoints(localeCfg)

	cfg := service.ConfigService.GetConfig()

	// Bind server address and port
	bindAddr := fmt.Sprintf("%v:%v", cfg.Http.Address, cfg.Http.Port)
	zap.L().Sugar().Info("server listens on ", bindAddr)

	// Configure and return the server
	server = &http.Server{Addr: bindAddr, Handler: engine}
	return server
}

// system initializing
func systemInit(cancelFunc context.CancelFunc, server *http.Server, ctx context.Context) *system.System {
	return system.Startup(ctx, &system.StartupParams{
		EnableEtcd:    false,
		EnableMongodb: true,
		EnableRedis:   true,
		Config:        service.ConfigService.GetConfig().GetServerConfig(),
		PreShutdown: func() error {
			//cancelFunc()
			return nil
		},
		PostShutdown: func() error {
			if server != nil {
				zap.S().Info("web server shuts down")
				if err := server.Shutdown(ctx); err != nil {
					zap.L().Error("unable to shut web server down", zap.Error(err))
					return err
				}
			}
			return nil
		},
	})
}
