package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris/v12"
	"time"
)

var mySigningKey = []byte("secret")

type User struct {
	Username    string
	Password    string
	Group       string
	Permissions []string
}

var Users = map[string]User{
	"user1": {Username: "user1", Password: "password1", Group: "admin", Permissions: []string{"read", "write"}},
	"user2": {Username: "user2", Password: "password2", Group: "user", Permissions: []string{"read"}},
}

func GenerateJWT(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 1).Unix(),
	})
	return token.SignedString(mySigningKey)
}

func JWTMiddleware(ctx iris.Context) {
	tokenString := ctx.GetHeader("Authorization")
	if tokenString == "" {
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Missing token"})
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.NewValidationError("unexpected signing method", jwt.ValidationErrorSignatureInvalid)
		}
		return mySigningKey, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username := claims["username"].(string)
		if user, exists := Users[username]; exists {
			ctx.Values().Set("user", user)
			ctx.Next()
		} else {
			ctx.StatusCode(iris.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "Invalid user"})
		}
	} else {
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": err.Error()})
	}
}

func PermissionMiddleware(requiredPermissions ...string) iris.Handler {
	return func(ctx iris.Context) {
		user := ctx.Values().Get("user").(User)
		for _, perm := range requiredPermissions {
			hasPermission := false
			for _, userPerm := range user.Permissions {
				if userPerm == perm {
					hasPermission = true
					break
				}
			}
			if !hasPermission {
				ctx.StatusCode(iris.StatusForbidden)
				ctx.JSON(iris.Map{"error": "You do not have permission to access this resource"})
				return
			}
		}
		ctx.Next()
	}
}
