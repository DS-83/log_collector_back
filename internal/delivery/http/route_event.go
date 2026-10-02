package http

import (
	"log_collect/internal"

	"github.com/gin-gonic/gin"
)

func RegisterRouteEventUser(r *gin.RouterGroup, uc internal.EventUseCase) {
	h := NewHandlerEvent(uc)

	r.GET("/getone/:id", h.GetByID)
	r.GET("/list", h.List)
}

func RegisterRouteEventApi(r *gin.RouterGroup, uc internal.EventUseCase) {
	h := NewHandlerEvent(uc)

	r.POST("/create", h.Create)
}
