package http

import (
	"errors"
	"log"
	e "log_collect/err"
	"log_collect/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	Error string `json:"error"`
}

// Helper function to handle bad request error
func handleBadRequest(c *gin.Context, err error, msg string) {
	c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	log.Printf("%s: %s", msg, err)
}

// Helper function to handle not acceptable error
func handleNotAcceptable(c *gin.Context, err error, msg string) {
	c.JSON(http.StatusNotAcceptable, errorResponse{Error: err.Error()})
	log.Printf("%s: %s", msg, err)
}

// Helper function to abort with internal server error and log the error
func abortWithError(c *gin.Context, err error, msg string) {
	c.AbortWithStatus(http.StatusInternalServerError)
	log.Printf("%s: %s", msg, err)
}

// Helper function to abort with error and log the error
func abortWithStatus(c *gin.Context, status int, msg string) {
	c.AbortWithStatus(status)
	log.Printf("%s: status: %d", msg, status)
}

// Helper function to handle unauthorized errors
func handleUnauthorized(c *gin.Context, err error, msg string) {
	c.JSON(http.StatusUnauthorized, errorResponse{Error: err.Error()})
	log.Printf("%s: %s", msg, err)
}

func handleError(c *gin.Context, err error, msg string) {
	switch {
	case errors.Is(err, e.ErrNotFound):
		c.Status(http.StatusNotFound)
	case errors.Is(err, e.ErrInvalidReqData):
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	log.Printf("%s: %s", msg, err)
}

func UserFromContext(c *gin.Context) *models.User {
	v, ok := c.Get(userContextKey)
	if !ok {
		return nil
	}
	return v.(*models.User)
}

func ApiKeyFromContext(c *gin.Context) *models.ApiKey {
	v, ok := c.Get(apiKeyContextKey)
	if !ok {
		return nil
	}
	return v.(*models.ApiKey)
}
