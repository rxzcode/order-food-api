package core

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error,omitempty"`
}

type SuccessResponse struct {
	Data interface{} `json:"data"`
}

func RespondError(c *gin.Context, status int, msg string, err error) {
	if err != nil {
		c.AbortWithStatusJSON(status, ErrorResponse{
			Message: msg,
			Error:   err.Error(),
		})
	} else {
		c.AbortWithStatusJSON(status, ErrorResponse{
			Message: msg,
		})
	}

}

func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{Data: data})
}
