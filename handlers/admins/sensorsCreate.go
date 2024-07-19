package admins

import (
	"errors"
	"fmt"
	"net/http"

	"admin-api/utils"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/db/models"
	"gorm.io/gorm"
)

// Sensor represents the structure of a sensor
type SensorReceived struct {
	Name     string    `json:"name"`
	Location string    `json:"location"`
	UserId   uuid.UUID `json:"userId"`
}

func ServeSensorsCreate(ctx iris.Context, db *gorm.DB) {
	var sensorReceived SensorReceived

	// Parse the request body
	if err := ctx.ReadJSON(&sensorReceived); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// Validate the sensor data
	if err := validateSensor(sensorReceived); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, err.Error(), fmt.Errorf("validateSensor error"))
		return
	}

	newSensor := models.Sensor{
		ID:       uuid.New(),
		UserID:   sensorReceived.UserId,
		Name:     sensorReceived.Name,
		Location: sensorReceived.Location,
		Secret:   uuid.New().String(),
	}
	if err := db.Create(&newSensor).Error; err != nil {
		utils.RespondError(ctx, http.StatusInternalServerError, "Invalid request", err)
		return
	}

	// Respond with the created sensor
	utils.RespondCreated(ctx, nil, "Sensor created successfully")
}

// validateSensor validates the sensor data
func validateSensor(sensor SensorReceived) error {
	if sensor.Name == "" {
		return errors.New("Sensor name is required")
	}
	if sensor.Location == "" {
		return errors.New("Sensor location is required")
	}
	return nil
}
