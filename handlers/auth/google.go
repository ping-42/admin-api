package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/db/models"
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
		_ = ctx.JSON(iris.Map{"error": "Invalid request"})
		return
	}

	// TODO mv to config
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		ctx.StatusCode(http.StatusInternalServerError)
		_ = ctx.JSON(iris.Map{"error": "Google client ID not configured"})
		return
	}

	payload, err := idtoken.Validate(context.Background(), body.Token, googleClientID)
	if err != nil {
		log.Printf("Failed to validate Google token: %v", err)
		ctx.StatusCode(http.StatusUnauthorized)
		_ = ctx.JSON(iris.Map{"error": "Invalid Google token"})
		return
	}

	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		log.Printf("Email not found in Google token payload: %v", payload.Claims)
		ctx.StatusCode(http.StatusUnauthorized)
		_ = ctx.JSON(iris.Map{"error": "Invalid Google token"})
		return
	}

	exp, ok := payload.Claims["exp"].(float64)
	if !ok || time.Now().Unix() > int64(exp) {
		log.Printf("Token is expired or exp claim not found")
		ctx.StatusCode(http.StatusUnauthorized)
		_ = ctx.JSON(iris.Map{"error": "Expired Google token"})
		return
	}

	user, err := getOrCreateGoogleUser(db, email)
	if err != nil {
		log.Printf("Failed to get or create user: %v", err)
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(iris.Map{"error": "Failed to generate user"})
		return
	}

	jwt, err := middleware.GenerateJWT(db, user)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(iris.Map{"error": "Failed to generate token"})
		return
	}

	_ = ctx.JSON(iris.Map{"token": jwt, "email": user.Email, "userGroupID": user.UserGroupID})
}

// getOrCreateGoogleUser checks if the user exists in the database, if not creates a new one.
func getOrCreateGoogleUser(db *gorm.DB, email string) (user models.User, err error) {
	if email == "" {
		err = fmt.Errorf("expected email address, got empty string")
		return
	}

	result := db.Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// User not found, create a new organization & admin user
		newOrg := models.Organization{
			ID:   uuid.New(),
			Name: "",
		}
		if err = db.Create(&newOrg).Error; err != nil {
			err = fmt.Errorf("error creating new organization: %v", err)
			return
		}

		newUser := models.User{
			ID:             uuid.New(),
			Email:          &email,
			UserGroupID:    2, // Admin
			OrganizationID: newOrg.ID,
			CreatedAt:      time.Now(),
			IsValidated:    true,
			IsActive:       true,
			LastLoginAt:    time.Now(),
		}
		if err = db.Create(&newUser).Error; err != nil {
			err = fmt.Errorf("error creating new user: %v", err)
			return
		}
		user = newUser

		log.Printf("New user created: %+v\n", newUser)
	} else if result.Error != nil {
		err = fmt.Errorf("error finding user: %v", result.Error)
		return
	} else {
		updateTx := db.Model(&models.User{}).Where("id = ?", user.ID).Update("last_login_at", time.Now())
		if updateTx.Error != nil {
			fmt.Println("Error updating user last_login_at", updateTx.Error) //TODO logger
			return
		}
		log.Printf("User found: %+v initiating email login\n", user.ID)
	}

	return
}
