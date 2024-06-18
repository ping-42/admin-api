package main

import (
	"github.com/ping-42/admin-api/data"
	"github.com/ping-42/admin-api/handlers"
	"github.com/ping-42/admin-api/middleware"

	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
)

func main() {
	app := iris.New()

	// CORS options
	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://reimagined-telegram-976gq4jxv69vcx6xv-3000.app.github.dev"}, // if running in codespaces needs the origin
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
	})

	// Apply CORS middleware
	app.UseRouter(crs)

	// Route to handle login
	app.Post("/login", handlers.LoginHandler)

	// Protected routes
	api := app.Party("/api", middleware.JWTMiddleware)
	{
		api.Get("/items", middleware.PermissionMiddleware("read_item"), func(ctx iris.Context) {
			ctx.JSON(data.Items)
		})
		api.Post("/items/update", middleware.PermissionMiddleware("write_item"), func(ctx iris.Context) {
			// Logic to add a new item
			ctx.JSON(iris.Map{"message": "Item added"})
		})
	}

	// Start the server
	app.Listen(":8081")
}
