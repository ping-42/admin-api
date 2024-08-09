package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/kataras/iris/v12"
	"github.com/ping-42/admin-api/middleware"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

// GoogleLoginHandler handles Google OAuth2 login.
func GoogleLoginHandler(ctx iris.Context, db *gorm.DB) {
	var body struct {
		Token string `json:"token"`
	}

	if err := ctx.ReadJSON(&body); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request"})
		return
	}

	// googleCredentials := os.Getenv("GOOGLE_CREDENTIALS") // TODO!!!!!!!

	// Validate the ID token and extract the user info.
	payload, err := idtoken.Validate(context.Background(), body.Token, "537878940617-g59al68c6s467ov8utt1jkcompevgriu.apps.googleusercontent.com")
	if err != nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid Google token"})
		return
	}

	// Extract the email from the token payload.
	email, ok := payload.Claims["email"].(string)
	if !ok {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid Google token"})
		return
	}

	// Check if the user exists in the database, or create a new one.
	user, err := getOrCreateUser(db, "", email)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to generate user"})
		return
	}

	// Generate a JWT token for your application.
	jwt, err := middleware.GenerateJWT(db, user)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(iris.Map{"token": jwt, "email": user.Email, "userGroupID": user.UserGroupID})
}

// User represents a user in your system.
type User struct {
	ID          uint
	Email       string
	UserGroupID uint
}

// AuthenticateGoogleUser checks if the user exists in the database.
func AuthenticateGoogleUser(email string, db *gorm.DB) (*User, error) {
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newUser := User{
				Email:       email,
				UserGroupID: 1, // Assign default group ID or create one dynamically
			}
			if err := db.Create(&newUser).Error; err != nil {
				return nil, err
			}
			return &newUser, nil
		}
		return nil, err
	}
	return &user, nil
}
