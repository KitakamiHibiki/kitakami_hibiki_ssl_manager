package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if domains == nil {
		domains = []store.Domain{}
	}
	c.JSON(http.StatusOK, domains)
}

func (h *DomainHandler) Create(c *gin.Context) {
	var d store.Domain
	if err := c.ShouldBindJSON(&d); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	d.UserID = c.GetUint("user_id")
	if d.Challenge == "" {
		d.Challenge = "http"
	}
	if err := h.db.CreateDomain(&d); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *DomainHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	userID := c.GetUint("user_id")
	role := c.GetString("role")
	if role != "admin" {
		if _, err := h.db.GetDomainByIDAndUser(uint(id), userID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "not your domain"})
			return
		}
	}

	if err := h.db.DeleteDomain(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
