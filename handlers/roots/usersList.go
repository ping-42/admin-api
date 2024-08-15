package roots

import (
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/db/models"
	"gorm.io/gorm"
)

type orgResponse struct {
	ID               uuid.UUID `json:"id"`
	OrganizationName string    `json:"organization_name"`
}

func ServeOrganizationsList(ctx iris.Context, db *gorm.DB) {

	var users []models.Organization
	if err := db.Find(&users).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(iris.Map{"error": "Failed to retrieve users"})
		return
	}

	var orgsResponse []orgResponse
	for _, s := range users {
		orgsResponse = append(orgsResponse, orgResponse{
			ID:               s.ID,
			OrganizationName: s.Name,
		})
	}

	_ = ctx.JSON(orgsResponse)
}
