package http

import (
	"log_collect/internal"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterRoutesUser(
	r *gin.RouterGroup,
	ac internal.UserUseCase,
	sessionTTL time.Duration,
	cookie cookieConfig,
) {
	h := NewHandlerUser(ac, sessionTTL, cookie)
	r.POST("/signin", h.SignIn)
}
