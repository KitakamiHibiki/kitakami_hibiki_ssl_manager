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
	domains, err := h.db.ListDomains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, domains)
}

func (h *DomainHandler) Create(c *gin.Context) {
	var d store.Domain
	if err := c.ShouldBindJSON(&d); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.db.DeleteDomain(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
