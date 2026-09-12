package response

import (
	"errors"
	"net/http"

	commonerrors "common-svr/internal/common/errors"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func Success(c *gin.Context, status int, data any) {
	c.JSON(status, Envelope{Code: "OK", Message: "ok", Data: data})
}

func Failure(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	message := "internal server error"

	switch {
	case errors.Is(err, commonerrors.ErrInvalidArgument):
		status = http.StatusBadRequest
		code = "INVALID_ARGUMENT"
		message = err.Error()
	case errors.Is(err, commonerrors.ErrNotFound):
		status = http.StatusNotFound
		code = "NOT_FOUND"
		message = err.Error()
	case errors.Is(err, commonerrors.ErrConflict):
		status = http.StatusConflict
		code = "CONFLICT"
		message = err.Error()
	}

	c.AbortWithStatusJSON(status, Envelope{Code: code, Message: message})
}
