package handler

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/deploy"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
)

type DeployHandler struct {
	db       *store.DB
	nginx    *deploy.NginxDeployer
	local    *deploy.LocalDeployer
	certDir  string
}

func NewDeployHandler(db *store.DB, nginx *deploy.NginxDeployer, local *deploy.LocalDeployer, certDir string) *DeployHandler {
	return &DeployHandler{db: db, nginx: nginx, local: local, certDir: certDir}
}

func (h *DeployHandler) Deploy(c *gin.Context) {
	var req struct {
		CertificateID uint   `json:"certificate_id"`
		Target        string `json:"target"` // nginx | local
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	cert, err := h.db.GetCertificateByIDAndUser(req.CertificateID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your certificate"})
		return
	}

	certPath := h.certDir + "/" + cert.Domain
	var deployErr error

	switch req.Target {
	case "nginx":
		deployErr = h.nginx.Deploy(cert.Domain, certPath)
	case "local":
		deployErr = h.local.Deploy(cert.Domain, certPath)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported target: " + req.Target})
		return
	}

	if deployErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": deployErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deployed"})
}

func (h *DeployHandler) Platform(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	})
}
