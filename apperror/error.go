package apperror

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorModel struct contains error message and http status codes.
type ErrorModel struct {
	Message string
	Code    int
}

var (
	// ErrCacheNotInitialized ...
	ErrCacheNotInitialized = errors.New("service not available at this moment, try after sometime")
)

func assertError(err error) *ErrorModel {
	if errors.Is(err, ErrCacheNotInitialized) {
		return &ErrorModel{
			Message: ErrCacheNotInitialized.Error(),
			Code:    http.StatusInternalServerError,
		}
	}
	return &ErrorModel{
		Message: "Unidentified Error",
		Code:    http.StatusInternalServerError,
	}
}

// ErrorResponse : Pass error message to the server.
func ErrorResponse(e error, c *gin.Context) {
	err := assertError(e)
	c.JSON(err.Code, gin.H{"message": err.Message})
	return
}
