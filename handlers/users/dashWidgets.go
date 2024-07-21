package users

import (
	"fmt"
	"strings"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/constants"
	"github.com/ping-42/42lib/db/models"
	"github.com/ping-42/admin-api/middleware"
	"github.com/ping-42/admin-api/utils"
	"gorm.io/gorm"
)

type countPerMonth struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type serveData struct {
	ActiveSensors  int             `json:"activeSensors"`
	EnabledSensors int64           `json:"enabledSensors"`
	TasksCompleted []countPerMonth `json:"tasksCompleted"`
}

func ServeDashWidgetData(ctx iris.Context, db *gorm.DB, redisClient *redis.Client) {
	userClaims, ok := ctx.Values().Get("UserClaims").(*middleware.UserClaims)
	if !ok {
		utils.RespondError(ctx, iris.StatusUnauthorized, "Unauthorized user", fmt.Errorf("ServeSensorsCreate casting to middleware.UserClaims error"))
		return
	}

	// get the count of enabled sensors
	var enabledSensors int64
	if err := db.Model(&models.Sensor{}).Where("user_id = ? AND is_active = ?", userClaims.UserId, true).Count(&enabledSensors).Error; err != nil {
		utils.RespondError(ctx, iris.StatusInternalServerError, "", err)
		return
	}

	//-----------------------
	// TODO: here we are getting the active/connected sensors per user this needs to be refactored
	var connectedSensorIDs []uuid.UUID
	connectedSensors, err := redisClient.Keys(constants.RedisActiveSensorsKeyPrefix + "*").Result()
	if err != nil {
		utils.RespondError(ctx, iris.StatusInternalServerError, "", err)
		return
	}
	for k, v := range connectedSensors {
		connectedSensors[k] = strings.TrimPrefix(v, constants.RedisActiveSensorsKeyPrefix)
		sensorID, er := uuid.Parse(connectedSensors[k])
		if er != nil {
			utils.RespondError(ctx, iris.StatusInternalServerError, "", er)
			return
		}
		connectedSensorIDs = append(connectedSensorIDs, sensorID)
	}
	// Fetch sensors from the database
	var sensors []models.Sensor
	if err := db.Where("id IN ? AND user_id=?", connectedSensorIDs, userClaims.UserId).Find(&sensors).Error; err != nil {
		utils.RespondError(ctx, iris.StatusInternalServerError, "", err)
		return
	}
	activeSensorsCount := len(sensors)
	//-----------------------

	// Query to get the count of tasks completed per month for the last 12 months for the specific user
	var tasksCompleted []countPerMonth
	var query = `
        WITH RECURSIVE last_12_months AS (
    SELECT date_trunc('month', current_date) - interval '11 months' AS month
    UNION ALL
    SELECT month + interval '1 month'
    FROM last_12_months
    WHERE month + interval '1 month' < date_trunc('month', current_date) + interval '1 month'
)
SELECT 
    TO_CHAR(last_12_months.month, 'Month') AS month, 
    COALESCE(COUNT(t.id), 0) AS count
FROM 
    last_12_months
LEFT JOIN 
    tasks t ON date_trunc('month', t.created_at) = last_12_months.month
LEFT JOIN 
    sensors s ON s.id = t.sensor_id
    AND s.user_id = ?
GROUP BY 
    last_12_months.month
ORDER BY 
    last_12_months.month;
    `

	if err := db.Raw(query, userClaims.UserId).Scan(&tasksCompleted).Error; err != nil {
		utils.RespondError(ctx, iris.StatusInternalServerError, "Failed to query active sensors", err)
		return
	}

	serveData := serveData{
		EnabledSensors: enabledSensors,
		ActiveSensors:  activeSensorsCount,
		TasksCompleted: tasksCompleted,
	}

	utils.RespondSuccess(ctx, serveData)
}

// Function to generate the last 12 months
// func getLast12Months() []string {
// 	var months []string
// 	for i := 11; i >= 0; i-- {
// 		month := time.Now().AddDate(0, -i, 0).Format("January")
// 		months = append(months, month)
// 	}
// 	return months
// }

//
// tasksCompletedData := []int{1, 18, 9, 17, 34, 22, 11, 15, 20, 25, 30, 35}
// "tasksCompletedData": tasksCompletedData,
// "tasksCompleted": 75,
//
// Prepare the response data
// var months []string
// var activeSensorsData []int
// for _, result := range results {
// 	months = append(months, result.Month)
// 	activeSensorsData = append(activeSensorsData, result.Count)
// }