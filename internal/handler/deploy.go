package handler

import (
	"runtime"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/deploy"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/response"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
)

type DeployHandler struct {
	db       *store.DB
	certDir  string
	proxyURL string
}

func NewDeployHandler(db *store.DB, certDir string, proxyURL string) *DeployHandler {
	return &DeployHandler{db: db, certDir: certDir, proxyURL: proxyURL}
}

func (h *DeployHandler) Platform(c *gin.Context) {
	response.OK(c, gin.H{
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"proxy_url": h.proxyURL,
	})
}

func (h *DeployHandler) Deploy(c *gin.Context) {
	var req struct {
		CertID uint `json:"cert_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.CertID == 0 {
		response.Error(c, 400, "cert_id is required")
		return
	}

	userID := c.GetUint("user_id")
	cert, err := h.db.GetCertificateByID(req.CertID)
	if err != nil {
		response.Error(c, 404, "certificate not found")
		return
	}
	if cert.UserID != userID {
		response.Error(c, 403, "not your certificate")
		return
	}
	if cert.Status != "issued" {
		response.Error(c, 400, "certificate not issued yet")
		return
	}
	if !cert.DeployEnabled || cert.DeployNodeID == 0 {
		response.Error(c, 400, "deploy not configured for this certificate")
		return
	}

	d := deploy.NewDeployer(h.db, h.certDir)
	go d.DeployCert(req.CertID)

	response.OK(c, gin.H{"message": "deploy started"})
}

func (h *DeployHandler) DeployLogs(c *gin.Context) {
	userID := c.GetUint("user_id")

	id, err := strconv.ParseUint(c.Query("cert_id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, 400, "cert_id is required")
		return
	}
	cert, err := h.db.GetCertificateByID(uint(id))
	if err != nil {
		response.Error(c, 404, "certificate not found")
		return
	}
	if cert.UserID != userID {
		response.Error(c, 403, "not your certificate")
		return
	}

	logs, err := h.db.ListDeployLogsByCert(uint(id))
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	if logs == nil {
		logs = []store.DeployLog{}
	}
	response.OK(c, logs)
}
