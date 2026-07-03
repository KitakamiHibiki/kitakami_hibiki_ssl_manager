package handler

import (
	"runtime"

	"github.com/gin-gonic/gin"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/response"
)

type DeployHandler struct{}

func NewDeployHandler() *DeployHandler {
	return &DeployHandler{}
}

func (h *DeployHandler) Platform(c *gin.Context) {
	response.OK(c, gin.H{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	})
}
