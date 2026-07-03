package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/response"
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
		response.Error(c, 500, "failed to load system config")
		return
	}
	response.OK(c, cfg)
}

func (h *SystemConfigHandler) Update(c *gin.Context) {
	cfg, err := h.db.GetSystemConfig()
	if err != nil {
		response.Error(c, 500, "failed to load system config")
		return
	}

	var req struct {
		JWTSecret *string `json:"jwt_secret"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request")
		return
	}

	if req.JWTSecret != nil {
		cfg.JWTSecret = *req.JWTSecret
	}

	if err := h.db.UpdateSystemConfig(cfg); err != nil {
		response.Error(c, 500, "failed to update system config")
		return
	}

	response.OK(c, gin.H{
		"message": "system config updated (some changes require restart)",
		"config":  cfg,
	})
}

func (h *SystemConfigHandler) Migrations(c *gin.Context) {
	list, err := h.db.AppliedMigrations()
	if err != nil {
		response.Error(c, 500, "failed to list migrations")
		return
	}
	if list == nil {
		list = []store.SchemaMigration{}
	}
	response.OK(c, list)
}
