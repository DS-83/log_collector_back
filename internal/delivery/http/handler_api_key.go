package http

import (
	"log_collect/internal"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HandlerApiKey struct {
	useCase internal.ApiKeyUseCase
}

func NewHandlerApiKey(uc internal.ApiKeyUseCase) *HandlerApiKey {
	return &HandlerApiKey{
		useCase: uc,
	}
}

type createApiKeyReq struct {
	Name   string `json:"name"   binding:"required"`
	Source string `json:"source" binding:"required"`
}

type createResp struct {
	Key string `json:"key"`
}

func (h *HandlerApiKey) Create(c *gin.Context) {
	in := new(createApiKeyReq)
	errMsg := "error create api key"

	if err := c.ShouldBindJSON(in); err != nil {
		handleBadRequest(c, err, errMsg)
		return
	}

	key, err := h.useCase.Create(c.Request.Context(), in.Name, in.Source)
	if err != nil {
		abortWithError(c, err, errMsg)
		return
	}

	out := createResp{Key: key}
	c.JSON(http.StatusCreated, out)
}
