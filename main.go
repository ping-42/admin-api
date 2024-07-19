package main

import (
	// "admin-api/handlers"

	"os"

	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	irisLogger "github.com/kataras/iris/v12/middleware/logger"
	"github.com/ping-42/42lib/config"
	"github.com/ping-42/42lib/db"
	"github.com/ping-42/42lib/logger"
	log "github.com/sirupsen/logrus"
)

// Release versioning magic
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var serverLogger = logger.Base("admin-api")

func init() {
	serverLogger.WithFields(log.Fields{
		"version":   version,
		"commit":    commit,
		"buildDate": date,
	}).Info("Starting PING42 Admin API Service...")
}

func main() {

	var serverLogger = logger.Base("server")
	serverLogger.WithFields(log.Fields{
		"version":   version,
		"commit":    commit,
		"buildDate": date,
	}).Info("Starting Admin API Server ...")

	configuration := config.GetConfig()
	gormClient, err := db.InitPostgreeDatabase(configuration.PostgreeDBDsn)
	if err != nil {
		serverLogger.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Unable to connect to Postgre Database")
		os.Exit(3)
	}

	redisClient, err := db.InitRedis(configuration.RedisHost, configuration.RedisPassword)
	if err != nil {
		serverLogger.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Unable to connect to Redis Database")
		os.Exit(4)
	}

	app := iris.New()

	// log all received requests
	app.Use(irisLogger.New())

	if config.CurrentEnv() == config.Dev {
		// CORS options
		crs := cors.New(cors.Options{
			AllowedOrigins:   []string{"http://localhost:3000", "https://reimagined-telegram-976gq4jxv69vcx6xv-3000.app.github.dev"},
			AllowCredentials: true,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
		})
		app.UseRouter(crs)
	}

	// Setup routes
	setupRoutes(app, gormClient, redisClient)

	// Start the server
	_ = app.Listen(":8081")
}
