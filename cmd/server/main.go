package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/acme"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/config"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/deploy"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/handler"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/middleware"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/platform"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/scheduler"
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
	db.SeedAdmin()

	certDir := plat.CertDir()
	if err := os.MkdirAll(certDir, 0755); err != nil {
		log.Fatalf("create cert dir: %v", err)
	}

	httpProvider := acme.NewHTTPProvider()
	authH := handler.NewAuthHandler(db, cfg.Auth.JWTSecret)
	domainH := handler.NewDomainHandler(db)
	certH := handler.NewCertHandler(db, cfg, certDir, httpProvider)
	userH := handler.NewUserHandler(db)
	nginxD := deploy.NewNginxDeployer(plat)
	localD := deploy.NewLocalDeployer(certDir)
	deployH := handler.NewDeployHandler(db, nginxD, localD, certDir)

	sched := scheduler.New(cfg, db)
	if err := sched.Start(func(domain, email, certDir string) error {
		return nil
	}); err != nil {
		log.Printf("[scheduler] start error: %v", err)
	}
	defer sched.Stop()

	r := gin.Default()

	r.GET("/.well-known/acme-challenge/*token", func(c *gin.Context) {
		httpProvider.ServeHTTP(c.Writer, c.Request)
	})

	authMw := middleware.AuthRequired(cfg.Auth.JWTSecret)
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
			domains.DELETE("/:id", adminMw, domainH.Delete)
		}

		certs := api.Group("/certs")
		certs.Use(authMw)
		{
			certs.POST("/apply", certH.Apply)
			certs.POST("/renew", certH.Renew)
			certs.GET("", certH.List)
			certs.GET("/:id", certH.Get)
			certs.GET("/:id/download", certH.Download)
		}

		deploy := api.Group("/deploy")
		deploy.Use(authMw, adminMw)
		{
			deploy.POST("", deployH.Deploy)
		}

		users := api.Group("/users")
		users.Use(authMw, adminMw)
		{
			users.GET("", userH.List)
			users.PUT("/:id", userH.UpdateRole)
			users.DELETE("/:id", userH.Delete)
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
