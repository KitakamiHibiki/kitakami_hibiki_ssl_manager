package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/config"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/handler"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/middleware"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/platform"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	plat := platform.Detect()

	dsn := cfg.Storage.DSN
	dbDir := filepath.Dir(dsn)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	db, err := store.InitDB(dsn)
	if err != nil {
		log.Fatalf("init db: %v", err)
	}
	if err := db.AfterInit(); err != nil {
		log.Fatalf("db init: %v", err)
	}
	if err := db.MarkIncompleteCertificatesAsError(); err != nil {
		log.Printf("[startup] mark incomplete certs: %v", err)
	}

	certDir := "./certs"
	if err := os.MkdirAll(certDir, 0755); err != nil {
		log.Fatalf("create cert dir: %v", err)
	}

	authH := handler.NewAuthHandler(db, cfg.Auth.JWTSecret, cfg.Auth.DeployKey)
	domainH := handler.NewDomainHandler(db)
	certH := handler.NewCertHandler(db, cfg, certDir)
	userH := handler.NewUserHandler(db)
	deployH := handler.NewDeployHandler(db, certDir)
	sysCfgH := handler.NewSystemConfigHandler(db)
	nodeH := handler.NewNodeHandler(db)

	r := gin.Default()

	authMw := middleware.AuthRequired(cfg.Auth.JWTSecret, cfg.Auth.DeployKey)
	adminMw := middleware.AdminRequired()

	api := r.Group("/api")
	{
		api.POST("/auth/register", authH.Register)
		api.POST("/auth/login", authH.Login)
		api.GET("/auth/me", authMw, authH.Me)

		domains := api.Group("/domains")
		domains.Use(authMw)
		{
			domains.GET("", domainH.List)
			domains.POST("", domainH.Create)
			domains.DELETE("", domainH.Delete)
			domains.PUT("", domainH.Update)
			domains.GET("/detail", domainH.Get)
		}

		certs := api.Group("/certs")
		certs.Use(authMw)
		{
			certs.GET("", certH.List)
			certs.POST("/apply", certH.Apply)
			certs.POST("/verify-dns", certH.VerifyDNS)
			certs.GET("/challenge-value", certH.ChallengeValue)
			certs.GET("/status", certH.Status)
			certs.GET("/detail", certH.Get)
			certs.GET("/download", certH.Download)
			certs.POST("/deploy", deployH.Deploy)
			certs.GET("/deploy-logs", deployH.DeployLogs)
		}

		users := api.Group("/users")
		users.Use(authMw, adminMw)
		{
			users.GET("", userH.List)
			users.PUT("", userH.UpdateRole)
			users.DELETE("", userH.Delete)
		}

		system := api.Group("/system")
		system.Use(authMw, adminMw)
		{
			system.GET("/config", sysCfgH.Get)
			system.PUT("/config", sysCfgH.Update)
			system.GET("/migrations", sysCfgH.Migrations)
		}

		nodes := api.Group("/nodes")
		nodes.Use(authMw)
		{
			nodes.GET("", nodeH.List)
			nodes.POST("", nodeH.Create)
			nodes.PUT("", nodeH.Update)
			nodes.DELETE("", nodeH.Delete)
			nodes.GET("/test", nodeH.Test)
		}

		api.GET("/platform", authMw, deployH.Platform)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
	})

	frontendDir := "web/dist"
	if _, err := os.Stat(frontendDir); err == nil {
		r.Use(func(c *gin.Context) {
			if c.Request.Method != "GET" || len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.Next()
				return
			}
			fs := http.FileServer(http.Dir(frontendDir))
			path := c.Request.URL.Path
			fullPath := filepath.Join(frontendDir, path)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				c.File(filepath.Join(frontendDir, "index.html"))
				c.Abort()
				return
			}
			fs.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		})
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("[server] starting on %s (platform: %s)", addr, plat)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
