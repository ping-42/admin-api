package roots

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/db/models"
	"github.com/ping-42/admin-api/utils"
	"gorm.io/gorm"
)

// Sensor represents the structure of a sensor
type SensorReceived struct {
	Name           string    `json:"name"`
	Location       string    `json:"location"`
	OrganizationId uuid.UUID `json:"organizationId"`
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
		ID:             uuid.New(),
		OrganizationID: sensorReceived.OrganizationId,
		Name:           sensorReceived.Name,
		Location:       sensorReceived.Location,
		Secret:         uuid.New().String(),
		IsActive:       true,
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
	if sensor.OrganizationId == uuid.Nil {
		return errors.New("Sensor organization is required")
	}
	if sensor.Name == "" {
		return errors.New("Sensor name is required")
	}
	if sensor.Location == "" {
		return errors.New("Sensor location is required")
	}
	return nil
}
