package http

import (
	"log_collect/internal"
	"net/http"

	"github.com/gin-gonic/gin"
)

const apiKeyContextKey = "api_key"

type AuthApiMiddleware struct {
	uc internal.ApiKeyUseCase
}

func NewAuthApiMiddleware(uc internal.ApiKeyUseCase) gin.HandlerFunc {
	return (&AuthApiMiddleware{
		uc: uc,
	}).Handle
}

func (m *AuthApiMiddleware) Handle(c *gin.Context) {
	raw := c.GetHeader("X-Api-Key")
	if raw == "" {
		abortWithStatus(c, http.StatusUnauthorized, "error api key middleware: missing key")
		return
	}
	key, err := m.uc.Authenticate(c.Request.Context(), raw)
	if err != nil {
		abortWithStatus(c, http.StatusUnauthorized, "error api key middleware: invalid key")
		return
	}
	c.Set(apiKeyContextKey, key)
	c.Next()
}
