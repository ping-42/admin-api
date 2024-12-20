package users

import (
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/constants"
	"github.com/ping-42/42lib/db/models"
	"github.com/ping-42/42lib/sensor"
	"github.com/ping-42/42lib/wss"
	"github.com/ping-42/admin-api/middleware"
	"github.com/ping-42/admin-api/utils"
	"gorm.io/gorm"
)

type sensorResponse struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	Location       string    `json:"location"`
	EnvToken       string    `json:"env_token"`
	IsConnected    bool      `json:"is_connected"`
	SensorVersion  string    `json:"sensor_version"`
}

func ServeSensorsList(ctx iris.Context, db *gorm.DB, redisClient *redis.Client) {
	userClaims, ok := ctx.Values().Get("UserClaims").(*middleware.UserClaims)
	if !ok {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(iris.Map{"error": "unauthorized user"})
		return
	}

	//-----------------------
	// TODO: here we are getting the connected/active sensors this needs to be moved to separate function
	var connectedSensorsData = make(map[uuid.UUID]wss.SensorConnection)
	connectedSensors, err := redisClient.Keys(constants.RedisActiveSensorsKeyPrefix + "*").Result()
	if err != nil {
		utils.RespondError(ctx, iris.StatusInternalServerError, "", err)
		return
	}
	for _, v := range connectedSensors {

		value, err := redisClient.Get(v).Result()
		if err != nil {
			utils.RespondError(ctx, iris.StatusInternalServerError, "", err)
			return
		}

		sensorConnection := wss.SensorConnection{}
		err = json.Unmarshal([]byte(value), &sensorConnection)
		if err != nil {
			utils.RespondError(ctx, iris.StatusInternalServerError, "", fmt.Errorf("Unmarshal sensorConnection: %v", err))
			return
		}
		connectedSensorsData[sensorConnection.SensorId] = sensorConnection
	}
	//-----------------------

	var sensors []models.Sensor
	if err := db.Select("id", "organization_id", "name", "location", "secret").Where("organization_id = ?", userClaims.OrganizationId).Find(&sensors).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(iris.Map{"error": "Failed to retrieve sensors"})
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
			_ = ctx.JSON(iris.Map{"error": "Failed to retrieve sensors"})
			return
		}

		sensorConnection, isConnected := connectedSensorsData[s.ID]

		sensorResponses = append(sensorResponses, sensorResponse{
			ID:             s.ID,
			OrganizationID: s.OrganizationID,
			Name:           s.Name,
			Location:       s.Location,
			EnvToken:       envToken,
			IsConnected:    isConnected,
			SensorVersion:  sensorConnection.SensorVersion,
		})
	}

	_ = ctx.JSON(sensorResponses)
}
