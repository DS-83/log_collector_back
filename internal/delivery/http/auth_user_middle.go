package http

import (
	"log_collect/internal"
	"net/http"

	"github.com/gin-gonic/gin"
)

const userContextKey = "user"

type AuthUserMiddleware struct {
	uc         internal.UserUseCase
	cookieName string
}

func NewAuthUserMiddleware(uc internal.UserUseCase, cookieName string) gin.HandlerFunc {
	return (&AuthUserMiddleware{
		uc:         uc,
		cookieName: cookieName,
	}).Handle
}

func (m *AuthUserMiddleware) Handle(c *gin.Context) {
	token, err := c.Cookie(m.cookieName)
	if err != nil {
		abortWithStatus(c, http.StatusUnauthorized, "error user middleware: missing session")
		return
	}
	user, err := m.uc.Authenticate(c.Request.Context(), token)
	if err != nil {
		abortWithStatus(c, http.StatusUnauthorized, "error user middleware: invalid token")
		return
	}

	c.Set(userContextKey, user)
	c.Next()
}
