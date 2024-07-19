package admins

import (
	"slices"
	"strings"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/constants"
	"github.com/ping-42/42lib/db/models"
	"github.com/ping-42/42lib/sensor"
	"github.com/ping-42/admin-api/utils"
	"gorm.io/gorm"
)

type sensorResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Location    string    `json:"location"`
	EnvToken    string    `json:"env_token"`
	IsConnected bool      `json:"is_connected"`
}

func ServeSensorsList(ctx iris.Context, db *gorm.DB, redisClient *redis.Client) {

	receivedUserId := ctx.Request().URL.Query().Get("user_id")

	//-----------------------
	// TODO: here we are getting the connected/active sensors this needs to be refactored
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
	//-----------------------

	var sensors []models.Sensor
	query := db.Select("id", "user_id", "name", "location", "secret")

	if receivedUserId != "" {
		query = query.Where("user_id = ?", receivedUserId)
	}

	if err := query.Find(&sensors).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve sensors"})
		return
	}

	var sensorResponses []sensorResponse
	for _, s := range sensors {

		sensorCreds := sensor.Creds{
			SensorId: s.ID,
			Secret:   s.Secret,
		}
		envToken, err := sensorCreds.GetSensorEnvToken()
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "Failed to retrieve sensors"})
			return
		}

		isConnected := slices.Contains(connectedSensorIDs, s.ID)

		sensorResponses = append(sensorResponses, sensorResponse{
			ID:          s.ID,
			UserID:      s.UserID,
			Name:        s.Name,
			Location:    s.Location,
			EnvToken:    envToken,
			IsConnected: isConnected,
		})
	}

	ctx.JSON(sensorResponses)
}
