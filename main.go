package main

import (
	"fmt"
	"os"

	"github.com/iris-contrib/middleware/cors"
	"github.com/jessevdk/go-flags"
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

// command-line options
type Options struct {
	Port     int    `short:"p" long:"port" description:"Port to listen on" default:"8081"`
	CorsHost string `short:"c" long:"corshost" description:"CORS Origin host to allow" default:"http://localhost:3000"`
}

func main() {

	var opts Options
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

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
			AllowedOrigins:   []string{opts.CorsHost},
			AllowCredentials: true,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
		})
		app.UseRouter(crs)
	}

	// Setup routes
	setupRoutes(app, gormClient, redisClient)

	// Start the server with the port from the flag
	err = app.Listen(fmt.Sprintf(":%d", opts.Port))
	if err != nil {
		serverLogger.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("app.Listen error")
		os.Exit(4)
	}
}
