package users

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/db/models"
	"github.com/ping-42/admin-api/middleware"
	"github.com/ping-42/admin-api/utils"
	"gorm.io/gorm"
)

type UserReceived struct {
	Email string `json:"email"`
}

func ServeOrganizationUsersCreate(ctx iris.Context, db *gorm.DB) {
	var userReceived UserReceived

	userClaims, ok := ctx.Values().Get("UserClaims").(*middleware.UserClaims)
	if !ok {
		utils.RespondError(ctx, iris.StatusUnauthorized, "Unauthorized user", fmt.Errorf("ServeUsersCreate casting to middleware.UserClaims error"))
		return
	}

	if err := ctx.ReadJSON(&userReceived); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, "Invalid request", err)
		return
	}

	if err := validateUser(userReceived); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, err.Error(), fmt.Errorf("validateSensor error"))
		return
	}

	newUser := models.User{
		ID:             uuid.New(),
		OrganizationID: userClaims.OrganizationId,
		Email:          &userReceived.Email,
		UserGroupID:    3, // User
		IsActive:       true,
		IsValidated:    false,
		CreatedAt:      time.Now(),
	}
	if err := db.Create(&newUser).Error; err != nil {
		utils.RespondError(ctx, http.StatusInternalServerError, "Invalid request", err)
		return
	}

	utils.RespondCreated(ctx, newUser, "New User created successfully")
}

func validateUser(user UserReceived) error {
	_, err := mail.ParseAddress(user.Email)
	if err != nil {
		return errors.New("valid email is required")
	}
	return nil
}
