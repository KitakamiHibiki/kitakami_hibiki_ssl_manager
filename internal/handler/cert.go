package handler

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/acme"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/config"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
)

type CertHandler struct {
	db       *store.DB
	cfg      *config.Config
	certDir  string
	provider *acme.HTTPProvider
}

func NewCertHandler(db *store.DB, cfg *config.Config, certDir string, provider *acme.HTTPProvider) *CertHandler {
	return &CertHandler{db: db, cfg: cfg, certDir: certDir, provider: provider}
}

func (h *CertHandler) Apply(c *gin.Context) {
	var req struct {
		DomainID uint   `json:"domain_id"`
		Domain   string `json:"domain"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.DomainID == 0 && req.Domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain_id or domain is required"})
		return
	}

	userID := c.GetUint("user_id")
	domain := req.Domain
	email := req.Email
	var domainID uint

	if req.DomainID > 0 {
		d, err := h.db.GetDomain(req.DomainID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
			return
		}
		if d.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "not your domain"})
			return
		}
		domain = d.Domain
		email = d.Email
		domainID = d.ID
	} else {
		if email == "" {
			email = "admin@" + domain
		}
		d := &store.Domain{Domain: domain, Email: email, Challenge: "http", UserID: userID}
		if err := h.db.CreateDomain(d); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		domainID = d.ID
	}

	certRecord := &store.Certificate{
		DomainID: domainID,
		Status:   "pending",
	}
	if err := h.db.CreateCertificate(certRecord); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go func() {
		client, err := acme.NewClient(email, h.cfg.ACME.Directory, h.certDir)
		if err != nil {
			log.Printf("[acme] new client error: %v", err)
			certRecord.Status = "error"
			h.db.UpdateCertificate(certRecord)
			return
		}

		if err := client.SetHTTPProvider(h.provider); err != nil {
			log.Printf("[acme] set http provider error: %v", err)
			certRecord.Status = "error"
			h.db.UpdateCertificate(certRecord)
			return
		}

		result, err := client.Obtain(domain)
		if err != nil {
			log.Printf("[acme] obtain error for %s: %v", domain, err)
			certRecord.Status = "error"
			h.db.UpdateCertificate(certRecord)
			return
		}

		certPEM := []byte(result.Certificate)
		expiry, _ := acme.ParseExpiry(certPEM)

		certRecord.Status = "issued"
		certRecord.IssuedAt = time.Now()
		certRecord.ExpiresAt = expiry
		h.db.UpdateCertificate(certRecord)
	}()

	c.JSON(http.StatusAccepted, certRecord)
}

func (h *CertHandler) Renew(c *gin.Context) {
	var req struct {
		CertificateID uint `json:"certificate_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	cert, err := h.db.GetCertificateByIDAndUser(req.CertificateID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your certificate"})
		return
	}

	domain, err := h.db.GetDomain(cert.DomainID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
		return
	}

	go func() {
		client, err := acme.NewClient(domain.Email, h.cfg.ACME.Directory, h.certDir)
		if err != nil {
			log.Printf("[acme] new client error: %v", err)
			return
		}
		if err := client.SetHTTPProvider(h.provider); err != nil {
			log.Printf("[acme] set http provider error: %v", err)
			return
		}
		result, err := client.Renew(cert.Domain)
		if err != nil {
			log.Printf("[acme] renew error for %s: %v", cert.Domain, err)
			return
		}
		certPEM := []byte(result.Certificate)
		expiry, _ := acme.ParseExpiry(certPEM)
		cert.Status = "issued"
		cert.IssuedAt = time.Now()
		cert.ExpiresAt = expiry
		h.db.UpdateCertificate(cert)
	}()
	c.JSON(http.StatusAccepted, gin.H{"message": "renew started"})
}

func (h *CertHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	role := c.GetString("role")

	var certs []store.Certificate
	var err error
	if role == "admin" {
		certs, err = h.db.ListCertificates()
	} else {
		certs, err = h.db.ListCertificatesByUser(userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if certs == nil {
		certs = []store.Certificate{}
	}
	c.JSON(http.StatusOK, certs)
}

func (h *CertHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	userID := c.GetUint("user_id")
	cert, err := h.db.GetCertificateByIDAndUser(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your certificate"})
		return
	}
	c.JSON(http.StatusOK, cert)
}

func (h *CertHandler) Download(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	userID := c.GetUint("user_id")
	cert, err := h.db.GetCertificateByIDAndUser(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your certificate"})
		return
	}

	certPath := h.certDir + "/" + cert.Domain + "/fullchain.pem"
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "cert file not found"})
		return
	}
	c.File(certPath)
}
