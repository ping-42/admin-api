package handlers

import (
	"github.com/kataras/iris/v12"
	"github.com/ping-42/admin-api/middleware"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginHandler(ctx iris.Context) {
	var req LoginRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(iris.Map{"error": "Invalid request"})
		return
	}

	// Validate user credentials
	if user, ok := middleware.Users[req.Username]; ok && user.Password == req.Password {
		token, err := middleware.GenerateJWT(req.Username)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(iris.Map{"error": "Failed to generate token"})
			return
		}
		_ = ctx.JSON(iris.Map{"token": token})
	} else {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(iris.Map{"error": "Invalid credentials"})
	}
}
