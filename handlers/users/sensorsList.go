package users

import (
	"slices"
	"strings"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/constants"
	"github.com/ping-42/42lib/db/models"
	"github.com/ping-42/42lib/sensor"
	"github.com/ping-42/admin-api/middleware"
	"github.com/ping-42/admin-api/utils"
	"gorm.io/gorm"
)

type sensorResponse struct {
	ID             uuid.UUID `json:"id"`
	OrganisationID uuid.UUID `json:"organisation_id"`
	Name           string    `json:"name"`
	Location       string    `json:"location"`
	EnvToken       string    `json:"env_token"`
	IsConnected    bool      `json:"is_connected"`
}

func ServeSensorsList(ctx iris.Context, db *gorm.DB, redisClient *redis.Client) {
	userClaims, ok := ctx.Values().Get("UserClaims").(*middleware.UserClaims)
	if !ok {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(iris.Map{"error": "unauthorized user"})
		return
	}

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
	if err := db.Select("id", "organisation_id", "name", "location", "secret").Where("organisation_id = ?", userClaims.OrganisationId).Find(&sensors).Error; err != nil {
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

		isConnected := slices.Contains(connectedSensorIDs, s.ID)

		sensorResponses = append(sensorResponses, sensorResponse{
			ID:             s.ID,
			OrganisationID: s.OrganisationID,
			Name:           s.Name,
			Location:       s.Location,
			EnvToken:       envToken,
			IsConnected:    isConnected,
		})
	}

	_ = ctx.JSON(sensorResponses)
}
