package users

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/admin-api/middleware"
	"github.com/ping-42/admin-api/utils"

	"gorm.io/gorm"
)

type res struct {
	Date     time.Time `gorm:"type:date"`
	SensorID uuid.UUID
	Count    int64
}

func ServeDashChartData(ctx iris.Context, db *gorm.DB) {
	userClaims, ok := ctx.Values().Get("UserClaims").(*middleware.UserClaims)
	if !ok {
		utils.RespondError(ctx, iris.StatusUnauthorized, "Unauthorized user", fmt.Errorf("ServeDashChartData: casting to middleware.UserClaims error"))
		return
	}

	// Calculate the date 30 days ago
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	// Construct the SQL query
	sqlQuery := `
		WITH dates AS (
			SELECT generate_series(?::date, ?::date, '1 day'::interval) as date
		)
		SELECT dates.date, COALESCE(tasks.sensor_id, ?) as sensor_id, COALESCE(COUNT(tasks.id), 0) as count
		FROM dates
		LEFT JOIN tasks ON tasks.created_at::date = dates.date
		AND tasks.task_status_id = ?
		LEFT JOIN sensors ON sensors.id = tasks.sensor_id
		AND sensors.organisation_id = ?
		GROUP BY dates.date, tasks.sensor_id
		ORDER BY dates.date, tasks.sensor_id
	`

	// Execute the raw SQL query
	var results []res
	err := db.Raw(sqlQuery, thirtyDaysAgo, time.Now(), uuid.Nil, 8, userClaims.OrganisationId).Scan(&results).Error
	if err != nil {
		utils.RespondError(ctx, iris.StatusInternalServerError, "Failed to query task counts", err)
		return
	}

	utils.RespondSuccess(ctx, results)
}
