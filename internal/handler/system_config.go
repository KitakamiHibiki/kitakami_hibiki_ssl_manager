package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
)

type SystemConfigHandler struct {
	db *store.DB
}

func NewSystemConfigHandler(db *store.DB) *SystemConfigHandler {
	return &SystemConfigHandler{db: db}
}

func (h *SystemConfigHandler) Get(c *gin.Context) {
	cfg, err := h.db.GetSystemConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load system config"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *SystemConfigHandler) Update(c *gin.Context) {
	cfg, err := h.db.GetSystemConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load system config"})
		return
	}

	var req struct {
		ACMEDirectory   *string `json:"acme_directory"`
		CheckInterval   *string `json:"check_interval"`
		RenewBeforeDays *int    `json:"renew_before_days"`
		NotifyEmail     *string `json:"notify_email"`
		NotifyWebhook   *string `json:"notify_webhook"`
		JWTSecret       *string `json:"jwt_secret"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.ACMEDirectory != nil {
		cfg.ACMEDirectory = *req.ACMEDirectory
	}
	if req.CheckInterval != nil {
		cfg.CheckInterval = *req.CheckInterval
	}
	if req.RenewBeforeDays != nil {
		cfg.RenewBeforeDays = *req.RenewBeforeDays
	}
	if req.NotifyEmail != nil {
		cfg.NotifyEmail = *req.NotifyEmail
	}
	if req.NotifyWebhook != nil {
		cfg.NotifyWebhook = *req.NotifyWebhook
	}
	if req.JWTSecret != nil {
		cfg.JWTSecret = *req.JWTSecret
	}

	// Warning if JWT secret changes — tokens signed with the old secret will be invalidated on restart.
	if err := h.db.UpdateSystemConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update system config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "system config updated (some changes require restart)",
		"config":   cfg,
	})
}

func (h *SystemConfigHandler) Migrations(c *gin.Context) {
	list, err := h.db.AppliedMigrations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list migrations"})
		return
	}
	if list == nil {
		list = []store.SchemaMigration{}
	}
	c.JSON(http.StatusOK, list)
}
