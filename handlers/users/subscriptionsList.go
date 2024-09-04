package users

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/db/models"
	"github.com/ping-42/admin-api/middleware"
	"gorm.io/gorm"
)

type subscriptionResponse struct {
	ID                     uint64        `json:"id"`
	TaskTypeName           string        `json:"task_type_name"`
	TestsCountSubscribed   int           `json:"tests_count_subscribed"`
	TestsCountExecuted     int           `json:"tests_count_executed"`
	Period                 time.Duration `json:"period"`
	Opts                   string        `json:"opts"`
	IsActive               bool          `json:"is_active"`
	LastExecutionCompleted time.Time     `json:"last_execution_completed"`
}

func ServeSubscriptionsList(ctx iris.Context, db *gorm.DB, redisClient *redis.Client) {

	userClaims, ok := ctx.Values().Get("UserClaims").(*middleware.UserClaims)
	if !ok {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(iris.Map{"error": "unauthorized user"})
		return
	}

	var subscriptions []models.Subscription
	if err := db.Preload("TaskType").Where("organization_id = ?", userClaims.OrganizationId).Find(&subscriptions).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(iris.Map{"error": "Failed to retrieve subscriptions"})
		return
	}

	var subscriptionResponses []subscriptionResponse
	for _, s := range subscriptions {
		subscriptionResponses = append(subscriptionResponses, subscriptionResponse{
			ID:                     s.ID,
			TaskTypeName:           s.TaskType.Type,
			TestsCountSubscribed:   s.TestsCountSubscribed,
			TestsCountExecuted:     s.TestsCountExecuted,
			Period:                 s.Period,
			Opts:                   string(s.Opts),
			IsActive:               s.IsActive,
			LastExecutionCompleted: s.LastExecutionCompleted,
		})
	}

	_ = ctx.JSON(subscriptionResponses)
}
