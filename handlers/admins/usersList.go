package admins

import (
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/db/models"
	"gorm.io/gorm"
)

type userResponse struct {
	ID            uuid.UUID `json:"id"`
	WalletAddress string    `json:"user_id"`
}

func ServeUsersList(ctx iris.Context, db *gorm.DB) {

	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve users"})
		return
	}

	var usersResponse []userResponse
	for _, s := range users {
		usersResponse = append(usersResponse, userResponse{
			ID:            s.ID,
			WalletAddress: s.WalletAddress,
		})
	}

	ctx.JSON(usersResponse)
}
