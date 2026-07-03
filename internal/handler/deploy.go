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
	db      *store.DB
	certDir string
}

func NewDeployHandler(db *store.DB, certDir string) *DeployHandler {
	return &DeployHandler{db: db, certDir: certDir}
}

func (h *DeployHandler) Platform(c *gin.Context) {
	response.OK(c, gin.H{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
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

	cert, err := h.db.GetCertificateByID(req.CertID)
	if err != nil {
		response.Error(c, 404, "certificate not found")
		return
	}
	if cert.Status != "issued" {
		response.Error(c, 400, "certificate not issued yet")
		return
	}

	userID := c.GetUint("user_id")
	dom, err := h.db.GetDomainByIDAndUser(cert.DomainID, userID)
	if err != nil {
		response.Error(c, 403, "not your domain")
		return
	}
	if !dom.DeployEnabled || dom.DeployNodeID == 0 {
		response.Error(c, 400, "deploy not configured for this domain")
		return
	}

	d := deploy.NewDeployer(h.db, h.certDir)
	go d.DeployCert(req.CertID)

	response.OK(c, gin.H{"message": "deploy started"})
}

func (h *DeployHandler) DeployLogs(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// prefer domain_id, fallback to cert_id
	if did, err := strconv.ParseUint(c.Query("domain_id"), 10, 64); err == nil && did > 0 {
		dom, err := h.db.GetDomainByIDAndUser(uint(did), userID)
		if err != nil || dom == nil {
			response.Error(c, 403, "not your domain")
			return
		}
		logs, total, err := h.db.ListDeployLogsByDomainPaginated(uint(did), offset, pageSize)
		if err != nil {
			response.Error(c, 500, err.Error())
			return
		}
		if logs == nil {
			logs = []store.DeployLog{}
		}
		response.OK(c, gin.H{"list": logs, "total": total})
		return
	}

	id, err := strconv.ParseUint(c.Query("cert_id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, 400, "cert_id or domain_id required")
		return
	}
	cert, err := h.db.GetCertificateByID(uint(id))
	if err != nil {
		response.Error(c, 404, "certificate not found")
		return
	}
	dom, err := h.db.GetDomainByIDAndUser(cert.DomainID, userID)
	if err != nil || dom == nil {
		response.Error(c, 403, "not your domain")
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
