package users

import (
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/db/models"
	"github.com/ping-42/admin-api/middleware"
	"gorm.io/gorm"
)

type orgUserResponse struct {
	UserId        uuid.UUID `json:"user_id"`
	WalletAddress string    `json:"wallet_address"`
	Email         string    `json:"email"`
	UserGroupID   uint64    `json:"user_group_id"`
	IsValidated   bool      `json:"is_validated"`
	LastLoginAt   time.Time `json:"last_login_at"`
}

func ServeOrganizationUserProfile(ctx iris.Context, db *gorm.DB) {

	userClaims, ok := ctx.Values().Get("UserClaims").(*middleware.UserClaims)
	if !ok {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(iris.Map{"error": "unauthorized user"})
		return
	}

	var user models.User
	if err := db.Where("organization_id=? and id=?", userClaims.OrganizationId, userClaims.UserId).Find(&user).Debug().Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(iris.Map{"error": "Failed to retrieve users"})
		return
	}

	var orgsResponse orgUserResponse

	walletAddress := ""
	email := ""
	if user.WalletAddress != nil {
		walletAddress = *user.WalletAddress
	}
	if user.Email != nil {
		email = *user.Email
	}

	orgsResponse = orgUserResponse{
		UserId:        user.ID,
		WalletAddress: walletAddress,
		Email:         email,
		UserGroupID:   user.UserGroupID,
		LastLoginAt:   user.LastLoginAt,
		IsValidated:   user.IsValidated,
	}

	_ = ctx.JSON(orgsResponse)
}
