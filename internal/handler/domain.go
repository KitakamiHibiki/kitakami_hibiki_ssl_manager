package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/response"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
)

type DomainHandler struct {
	db *store.DB
}

func NewDomainHandler(db *store.DB) *DomainHandler {
	return &DomainHandler{db: db}
}

func (h *DomainHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	role := c.GetString("role")

	var domains []store.Domain
	var err error
	if role == "admin" {
		domains, err = h.db.ListDomains()
	} else {
		domains, err = h.db.ListDomainsByUser(userID)
	}
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	if domains == nil {
		domains = []store.Domain{}
	}
	response.OK(c, domains)
}

func (h *DomainHandler) Create(c *gin.Context) {
	var d store.Domain
	if err := c.ShouldBindJSON(&d); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	d.UserID = c.GetUint("user_id")
	if err := h.db.CreateDomain(&d); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, d)
}

func (h *DomainHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, 400, "invalid id")
		return
	}

	userID := c.GetUint("user_id")
	role := c.GetString("role")
	if role != "admin" {
		if _, err := h.db.GetDomainByIDAndUser(uint(id), userID); err != nil {
			response.Error(c, 403, "not your domain")
			return
		}
	}

	if err := h.db.DeleteDomain(uint(id)); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "deleted"})
}

func (h *DomainHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, 400, "invalid id")
		return
	}

	userID := c.GetUint("user_id")
	role := c.GetString("role")

	var domain *store.Domain
	if role == "admin" {
		domain, err = h.db.GetDomain(uint(id))
	} else {
		domain, err = h.db.GetDomainByIDAndUser(uint(id), userID)
	}
	if err != nil {
		response.Error(c, 404, "domain not found")
		return
	}

	response.OK(c, gin.H{"domain": domain})
}

func (h *DomainHandler) Update(c *gin.Context) {
	var req struct {
		ID            uint   `json:"id"`
		DeployEnabled bool   `json:"deploy_enabled"`
		DeployNodeID  uint   `json:"deploy_node_id"`
		DeployType    string `json:"deploy_type"`
		CertName      string `json:"cert_name"`
		CertPath      string `json:"cert_path"`
		KeyName       string `json:"key_name"`
		KeyPath       string `json:"key_path"`
		AutoRenew     bool   `json:"auto_renew"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == 0 {
		response.Error(c, 400, "invalid request")
		return
	}

	userID := c.GetUint("user_id")
	role := c.GetString("role")

	var domain *store.Domain
	var err error
	if role == "admin" {
		domain, err = h.db.GetDomain(req.ID)
	} else {
		domain, err = h.db.GetDomainByIDAndUser(req.ID, userID)
	}
	if err != nil {
		response.Error(c, 404, "domain not found")
		return
	}

	domain.DeployEnabled = req.DeployEnabled
	domain.DeployNodeID = req.DeployNodeID
	domain.DeployType = req.DeployType
	domain.CertName = req.CertName
	domain.CertPath = req.CertPath
	domain.KeyName = req.KeyName
	domain.KeyPath = req.KeyPath
	domain.AutoRenew = req.AutoRenew
	if err := h.db.UpdateDomain(domain); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, domain)
}
