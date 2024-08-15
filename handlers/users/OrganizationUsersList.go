package users

import (
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/db/models"
	"github.com/ping-42/admin-api/middleware"
	"gorm.io/gorm"
)

type orgUsersResponse struct {
	UserId        uuid.UUID `json:"user_id"`
	WalletAddress string    `json:"wallet_address"`
	Email         string    `json:"email"`
	UserGroupID   uint64    `json:"user_group_id"`
}

func ServeOrganizationUsersList(ctx iris.Context, db *gorm.DB) {

	userClaims, ok := ctx.Values().Get("UserClaims").(*middleware.UserClaims)
	if !ok {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(iris.Map{"error": "unauthorized user"})
		return
	}

	var users []models.User
	if err := db.Where("organization_id=?", userClaims.OrganizationId).Find(&users).Debug().Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(iris.Map{"error": "Failed to retrieve users"})
		return
	}

	var orgsResponse []orgUsersResponse
	for _, s := range users {

		// for safe dereferencing
		walletAddress := ""
		email := ""
		if s.WalletAddress != nil {
			walletAddress = *s.WalletAddress
		}
		if s.Email != nil {
			email = *s.Email
		}

		orgsResponse = append(orgsResponse, orgUsersResponse{
			UserId:        s.ID,
			WalletAddress: walletAddress,
			Email:         email,
			UserGroupID:   s.UserGroupID,
		})
	}

	_ = ctx.JSON(orgsResponse)
}
