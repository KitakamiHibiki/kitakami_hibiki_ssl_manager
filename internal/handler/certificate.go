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

	domainID, _ := strconv.ParseUint(c.Query("domain_id"), 10, 64)

	var certs []store.Certificate
	var total int64
	var err error

	if role == "admin" {
		if domainID > 0 {
			certs, err = h.db.GetCertificatesByDomainID(uint(domainID))
			total = int64(len(certs))
		} else {
			certs, total, err = h.db.ListAllCertificates(offset, pageSize)
		}
	} else {
		if domainID > 0 {
			certs, total, err = h.db.ListCertificatesByDomainAndUser(uint(domainID), userID, offset, pageSize)
		} else {
			certs, total, err = h.db.ListCertificatesByUser(userID, offset, pageSize)
		}
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
