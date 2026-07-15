package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/response"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
)

type CertificateHandler struct {
	db *store.DB
}

func NewCertificateHandler(db *store.DB) *CertificateHandler {
	return &CertificateHandler{db: db}
}

func (h *CertificateHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	role := c.GetString("role")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var certs []store.Certificate
	var total int64
	var err error

	if role == "admin" {
		certs, total, err = h.db.ListAllCertificates(offset, pageSize)
	} else {
		certs, total, err = h.db.ListCertificatesByUser(userID, offset, pageSize)
	}
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	if certs == nil {
		certs = []store.Certificate{}
	}
	response.OK(c, gin.H{
		"list":      certs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *CertificateHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, 400, "invalid id")
		return
	}

	userID := c.GetUint("user_id")
	role := c.GetString("role")

	var cert *store.Certificate
	if role == "admin" {
		cert, err = h.db.GetCertificateByID(uint(id))
	} else {
		cert, err = h.db.GetCertificateByIDAndUser(uint(id), userID)
	}
	if err != nil {
		response.Error(c, 404, "certificate not found")
		return
	}

	response.OK(c, cert)
}

func (h *CertificateHandler) Update(c *gin.Context) {
	var req struct {
		ID            uint   `json:"id"`
		DeployEnabled *bool  `json:"deploy_enabled"`
		DeployNodeID  *uint  `json:"deploy_node_id"`
		DeployType    string `json:"deploy_type"`
		CertName      string `json:"cert_name"`
		CertPath      string `json:"cert_path"`
		KeyName       string `json:"key_name"`
		KeyPath       string `json:"key_path"`
		AutoRenew     *bool  `json:"auto_renew"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == 0 {
		response.Error(c, 400, "invalid request")
		return
	}

	userID := c.GetUint("user_id")
	role := c.GetString("role")

	cert, err := h.db.GetCertificateByID(req.ID)
	if err != nil {
		response.Error(c, 404, "not found")
		return
	}
	if role != "admin" && cert.UserID != userID {
		response.Error(c, 403, "not your certificate")
		return
	}

	updates := make(map[string]interface{})
	if req.DeployEnabled != nil {
		updates["deploy_enabled"] = *req.DeployEnabled
	}
	if req.DeployNodeID != nil {
		updates["deploy_node_id"] = *req.DeployNodeID
	}
	if req.DeployType != "" {
		updates["deploy_type"] = req.DeployType
	}
	if req.CertName != "" {
		updates["cert_name"] = req.CertName
	}
	if req.CertPath != "" {
		updates["cert_path"] = req.CertPath
	}
	if req.KeyName != "" {
		updates["key_name"] = req.KeyName
	}
	if req.KeyPath != "" {
		updates["key_path"] = req.KeyPath
	}
	if req.AutoRenew != nil {
		updates["auto_renew"] = *req.AutoRenew
	}

	if err := h.db.UpdateCertificate(req.ID, updates); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "updated"})
}

func (h *CertificateHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, 400, "invalid id")
		return
	}

	userID := c.GetUint("user_id")
	role := c.GetString("role")

	if role == "admin" {
		if err := h.db.DeleteCertificate(uint(id)); err != nil {
			response.Error(c, 500, err.Error())
			return
		}
	} else {
		if err := h.db.DeleteCertificateByIDAndUser(uint(id), userID); err != nil {
			response.Error(c, 403, "not your certificate")
			return
		}
	}

	response.OK(c, gin.H{"message": "deleted"})
}
