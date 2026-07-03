package handler

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/acme"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/config"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/response"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
)

func lookupTXT(domain string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		"Resolve-DnsName -Name '"+domain+"' -Type TXT -Server 8.8.8.8 | Select-Object -ExpandProperty Strings")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result, nil
}

type CertHandler struct {
	db      *store.DB
	cfg     *config.Config
	certDir string
}

func NewCertHandler(db *store.DB, cfg *config.Config, certDir string) *CertHandler {
	return &CertHandler{db: db, cfg: cfg, certDir: certDir}
}

func (h *CertHandler) Apply(c *gin.Context) {
	var req struct {
		DomainID uint `json:"domain_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.DomainID == 0 {
		response.Error(c, 400, "domain_id is required")
		return
	}

	userID := c.GetUint("user_id")
	d, err := h.db.GetDomain(req.DomainID)
	if err != nil {
		response.Error(c, 404, "domain not found")
		return
	}
	if d.UserID != userID {
		response.Error(c, 403, "not your domain")
		return
	}

	app := &acme.Application{
		Domain:   d.Domain,
		DomainID: d.ID,
		Status:   "verifying",
	}
	acme.SetApplication(d.Domain, app)

	go func() {
		client, err := acme.NewClient(d.Email, h.cfg.ACME.Directory, h.certDir)
		if err != nil {
			log.Printf("[acme] new client error: %v", err)
			app.Status = "error"
			app.ErrorMsg = acme.CleanError(err)
			return
		}

		client.SetDNSProvider(h.cfg.ACME.RecursiveNameservers)

		result, err := client.Obtain(d.Domain)
		if err != nil {
			log.Printf("[acme] obtain error for %s: %v", d.Domain, err)
			app.Status = "error"
			app.ErrorMsg = acme.CleanError(err)
			if app.CertID > 0 {
				h.db.UpdateCertificate(&store.Certificate{ID: app.CertID, Status: "error", ErrorMsg: acme.CleanError(err)})
			}
			return
		}

		certPEM := []byte(result.Certificate)
		expiry, _ := acme.ParseExpiry(certPEM)

		certPath := h.certDir + "/" + d.Domain
		os.MkdirAll(certPath, 0755)
		os.WriteFile(certPath+"/fullchain.pem", certPEM, 0644)
		os.WriteFile(certPath+"/privkey.pem", result.PrivateKey, 0600)

		if app.CertID > 0 {
			certRecord := &store.Certificate{ID: app.CertID}
			certRecord.Status = "issued"
			certRecord.IssuedAt = time.Now()
			certRecord.ExpiresAt = expiry
			h.db.UpdateCertificate(certRecord)
		} else {
			certRecord := &store.Certificate{
				DomainID:  d.ID,
				Status:    "issued",
				IssuedAt:  time.Now(),
				ExpiresAt: expiry,
			}
			h.db.CreateCertificate(certRecord)
			app.CertID = certRecord.ID
		}
		app.Status = "issued"
		app.ExpiresAt = expiry
	}()

	response.OK(c, gin.H{"domain": d.Domain, "status": "verifying"})
}

func (h *CertHandler) VerifyDNS(c *gin.Context) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Domain == "" {
		response.Error(c, 400, "domain is required")
		return
	}

	app := acme.GetApplication(req.Domain)
	if app == nil {
		response.Error(c, 404, "no pending application")
		return
	}

	names, err := lookupTXT("_acme-challenge." + req.Domain)
	if err != nil {
		log.Printf("[dns] lookup TXT error for %s: %v", req.Domain, err)
	}
	if err != nil || len(names) == 0 {
		response.Error(c, 400, "DNS TXT 记录未找到，请确认已添加")
		return
	}

	certRecord := &store.Certificate{
		DomainID: app.DomainID,
		Status:   "issuing",
	}
	if err := h.db.CreateCertificate(certRecord); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	app.Status = "issuing"
	app.CertID = certRecord.ID
	response.OK(c, gin.H{"status": "issuing", "cert_id": certRecord.ID})
}

func (h *CertHandler) Status(c *gin.Context) {
	domain := c.Query("domain")
	if domain == "" {
		response.Error(c, 400, "domain is required")
		return
	}
	app := acme.GetApplication(domain)
	if app == nil {
		response.Error(c, 404, "no pending application")
		return
	}
	response.OK(c, app)
}

func (h *CertHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	certs, total, err := h.db.ListCertificatesByUser(userID, offset, pageSize)
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

func (h *CertHandler) ChallengeValue(c *gin.Context) {
	domain := c.Query("domain")
	if domain == "" {
		response.Error(c, 400, "domain is required")
		return
	}
	val := acme.ManualDNS.GetKeyAuth(domain)
	if val == "" {
		response.Error(c, 404, "no pending challenge for this domain")
		return
	}
	response.OK(c, gin.H{"challenge_value": val})
}

func (h *CertHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, 400, "invalid id")
		return
	}
	userID := c.GetUint("user_id")
	cert, err := h.db.GetCertificateByIDAndUser(uint(id), userID)
	if err != nil {
		response.Error(c, 404, "not found")
		return
	}
	response.OK(c, cert)
}

func (h *CertHandler) Download(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, 400, "invalid id")
		return
	}
	fileType := c.Query("type")
	if fileType != "fullchain" && fileType != "privkey" {
		response.Error(c, 400, "type must be fullchain or privkey")
		return
	}

	userID := c.GetUint("user_id")
	cert, err := h.db.GetCertificateByIDAndUser(uint(id), userID)
	if err != nil {
		response.Error(c, 404, "not found")
		return
	}
	if cert.Status != "issued" {
		response.Error(c, 400, "certificate not issued yet")
		return
	}

	filePath := filepath.Join(h.certDir, cert.Domain, fileType+".pem")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		response.Error(c, 404, "file not found")
		return
	}

	c.FileAttachment(filePath, cert.Domain+"."+fileType+".pem")
}
