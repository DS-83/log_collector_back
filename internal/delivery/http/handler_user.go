package http

import (
	"log_collect/internal"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type cookieConfig struct {
	name     string
	path     string
	domain   string
	secure   bool
	httpOnly bool
}

func NewCookieConfig(name, path, domain string, secure, httpOnly bool) cookieConfig {
	return cookieConfig{name, path, domain, secure, httpOnly}
}

type HandlerUser struct {
	authUc     internal.UserUseCase
	sessionTTL time.Duration
	cookie     cookieConfig
}

func NewHandlerUser(uc internal.UserUseCase, sessionTTL time.Duration, cookie cookieConfig) *HandlerUser {
	return &HandlerUser{
		authUc:     uc,
		sessionTTL: sessionTTL,
		cookie:     cookie,
	}
}

type signInInput struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *HandlerUser) SignIn(c *gin.Context) {
	errMsg := "error sign-in"
	in := new(signInInput)

	if err := c.ShouldBindJSON(in); err != nil {
		handleBadRequest(c, err, errMsg)
		return
	}

	token, err := h.authUc.Login(c.Request.Context(), in.Login, in.Password)
	if err != nil {
		handleUnauthorized(c, err, errMsg)
		return
	}

	c.SetCookie(
		h.cookie.name,
		token,
		int(h.sessionTTL.Seconds()),
		h.cookie.path,
		h.cookie.domain,
		h.cookie.secure,
		h.cookie.httpOnly,
	)

	c.Status(http.StatusOK)
}
