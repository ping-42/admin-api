package main

import (
	"github.com/go-redis/redis"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/admin-api/handlers/admins"
	"github.com/ping-42/admin-api/handlers/auth"
	"github.com/ping-42/admin-api/handlers/users"
	"github.com/ping-42/admin-api/middleware"
	"gorm.io/gorm"
)

// SetupRoutes initializes all the routes and handlers
func setupRoutes(app *iris.Application, db *gorm.DB, redisClient *redis.Client) {
	// Public routes
	app.Post("/login/nonce", func(ctx iris.Context) {
		auth.NonceHandler(ctx, db)
	})
	app.Post("/login/init", func(ctx iris.Context) {
		auth.LoginHandler(ctx, db)
	})

	// Admin routes
	apiRoutesAdmin := app.Party("/api/admin", middleware.ValidateJWTMiddleware, middleware.ValidateAdminMiddleware)
	{
		apiRoutesAdmin.Get("/sensors/list", middleware.PermissionMiddleware("read"), func(ctx iris.Context) {
			admins.ServeSensorsList(ctx, db, redisClient)
		})
		apiRoutesAdmin.Post("/sensors/create", middleware.PermissionMiddleware("create"), func(ctx iris.Context) {
			admins.ServeSensorsCreate(ctx, db)
		})
		apiRoutesAdmin.Get("/users/list", middleware.PermissionMiddleware("read"), func(ctx iris.Context) {
			admins.ServeUsersList(ctx, db)
		})
	}

	// User routes
	apiRoutes := app.Party("/api", middleware.ValidateJWTMiddleware)
	{
		// NOTE: If needed we can create more concreate permissions e.g. read_dash, sensors_create

		apiRoutes.Get("/dash-widgets-data", middleware.PermissionMiddleware("read"), func(ctx iris.Context) {
			users.ServeDashWidgetData(ctx, db, redisClient)
		})
		apiRoutes.Get("/dash-chart-data", middleware.PermissionMiddleware("read"), func(ctx iris.Context) {
			users.ServeDashChartData(ctx, db)
		})
		apiRoutes.Get("/sensors/list", middleware.PermissionMiddleware("read"), func(ctx iris.Context) {
			users.ServeSensorsList(ctx, db, redisClient)
		})
		apiRoutes.Post("/sensors/create", middleware.PermissionMiddleware("create"), func(ctx iris.Context) {
			users.ServeSensorsCreate(ctx, db)
		})
	}
}