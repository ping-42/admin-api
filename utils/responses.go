package utils

import (
	"fmt"
	"net/http"

	"github.com/kataras/iris/v12"
)

// SuccessResponse represents a standard structure for successful responses
type SuccessResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse represents a standard structure for error responses
type ErrorResponse struct {
	Status  string `json:"status"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// RespondSuccess sends a successful response
func RespondSuccess(ctx iris.Context, data interface{}) {
	response := SuccessResponse{
		Status: "success",
		Data:   data,
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(response)
}

// RespondCreated sends a resource created response
func RespondCreated(ctx iris.Context, data interface{}, clientMsg string) {
	response := SuccessResponse{
		Status:  "success",
		Data:    data,
		Message: clientMsg,
	}
	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(response)
}

// RespondError sends an error response
func RespondError(ctx iris.Context, statusCode int, clientMsg string, logError error) {
	response := ErrorResponse{
		Status:  "error",
		Message: clientMsg,
	}

	fmt.Printf("Response error:%v, %v, %v\n", ctx.Request().RequestURI, clientMsg, logError.Error()) //TODO use logger

	ctx.StatusCode(statusCode)
	ctx.JSON(response)
}
