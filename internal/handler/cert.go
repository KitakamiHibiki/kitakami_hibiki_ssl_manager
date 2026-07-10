package handler

import (
	"context"
	"encoding/json"
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
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/deploy"
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

// extractRootDomain returns the root domain from a domain name.
// e.g. "www.example.com" -> "example.com", "*.example.com" -> "example.com"
func extractRootDomain(domain string) string {
	domain = strings.TrimPrefix(domain, "*.")
	parts := strings.Split(domain, ".")
	if len(parts) <= 2 {
		return domain
	}
	return strings.Join(parts[len(parts)-2:], ".")
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
		DomainID     uint     `json:"domain_id"`
		ExtraDomains []string `json:"extra_domains"`
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

	// Build full domain list: primary + extras
	allDomains := []string{d.Domain}
	allDomains = append(allDomains, req.ExtraDomains...)

	// Validate extras share the same root domain
	root := extractRootDomain(d.Domain)
	for _, extra := range req.ExtraDomains {
		if extractRootDomain(extra) != root {
			response.Error(c, 400, "extra domain '"+extra+"' does not share the same root domain as '"+d.Domain+"'")
			return
		}
	}

	// Deduplicate
	seen := map[string]bool{}
	unique := allDomains[:0]
	for _, name := range allDomains {
		if !seen[name] {
			seen[name] = true
			unique = append(unique, name)
		}
	}
	allDomains = unique

	domainsJSON, _ := json.Marshal(allDomains)

	certRecord := &store.Certificate{
		DomainID: d.ID,
		Domains:  string(domainsJSON),
		Status:   "verifying",
	}
	if err := h.db.CreateCertificate(certRecord); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	app := &acme.Application{
		Domain:   d.Domain,
		DomainID: d.ID,
		Domains:  allDomains,
		Status:   "verifying",
		CertID:   certRecord.ID,
	}
	acme.SetApplication(d.Domain, app)

	go func() {
		client, err := acme.NewClient(d.Email, h.cfg.ACME.Directory, h.certDir)
		if err != nil {
			log.Printf("[acme] new client error: %v", err)
			app.Status = "error"
			app.ErrorMsg = acme.CleanError(err)
			h.db.UpdateCertificate(&store.Certificate{ID: app.CertID, Status: "error", ErrorMsg: acme.CleanError(err)})
			return
		}

		client.SetDNSProvider(h.cfg.ACME.RecursiveNameservers)

		result, err := client.Obtain(allDomains)
		if err != nil {
			log.Printf("[acme] obtain error for %s: %v", d.Domain, err)
			app.Status = "error"
			app.ErrorMsg = acme.CleanError(err)
			h.db.UpdateCertificate(&store.Certificate{ID: app.CertID, Status: "error", ErrorMsg: acme.CleanError(err)})
			return
		}

		certPEM := []byte(result.Certificate)
		expiry, _ := acme.ParseExpiry(certPEM)

		certPath := h.certDir + "/" + d.Domain
		os.MkdirAll(certPath, 0755)
		os.WriteFile(certPath+"/fullchain.pem", certPEM, 0644)
		os.WriteFile(certPath+"/privkey.pem", result.PrivateKey, 0600)

		certRecord := &store.Certificate{ID: app.CertID}
		certRecord.Status = "issued"
		certRecord.IssuedAt = time.Now()
		certRecord.ExpiresAt = expiry
		h.db.UpdateCertificate(certRecord)
		app.Status = "issued"
		app.ExpiresAt = expiry

		// auto-deploy
		d2, _ := h.db.GetDomain(d.ID)
		if d2 != nil && d2.DeployEnabled && d2.DeployNodeID > 0 {
			deployer := deploy.NewDeployer(h.db, h.certDir)
			go deployer.DeployCert(app.CertID)
		}
	}()

	response.OK(c, gin.H{"domain": d.Domain, "cert_id": certRecord.ID, "domains": allDomains, "status": "verifying"})
}

func (h *CertHandler) VerifyDNS(c *gin.Context) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Domain == "" {
		response.Error(c, 400, "domain is required")
		return
	}

	// Find the application by any of the SAN domains
	var app *acme.Application
	for _, a := range getAllApps() {
		for _, d := range a.Domains {
			if d == req.Domain || strings.TrimPrefix(d, "*.") == strings.TrimPrefix(req.Domain, "*.") {
				app = a
				break
			}
		}
		if app != nil {
			break
		}
	}
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

	expected := acme.ManualDNS.GetKeyAuth(req.Domain)
	if expected == "" {
		response.Error(c, 400, "no pending challenge for this domain")
		return
	}
	matched := false
	for _, n := range names {
		if strings.Contains(n, expected) {
			matched = true
			break
		}
	}
	if !matched {
		response.Error(c, 400, "DNS TXT 记录值不匹配，请确认已更新为最新挑战值")
		return
	}

	// Signal this specific domain
	acme.ManualDNS.Signal(req.Domain)

	// Check if all domains are now signaled
	allDone := true
	var pendingList []string
	for _, d := range app.Domains {
		if !acme.ManualDNS.IsSignaled(d) {
			allDone = false
			pendingList = append(pendingList, d)
		}
	}

	if allDone {
		h.db.UpdateCertificate(&store.Certificate{ID: app.CertID, Status: "issuing"})
		app.Status = "issuing"
		response.OK(c, gin.H{"status": "issuing", "cert_id": app.CertID, "all_verified": true})
	} else {
		response.OK(c, gin.H{"status": "verifying", "cert_id": app.CertID, "all_verified": false, "pending": pendingList})
	}
}

func getAllApps() []*acme.Application {
	// Access the internal map via a helper on ManualProvider
	// We use a trick: iterate through known domains
	// Actually, we need a way to list all apps. Let's add a method.
	return acme.GetAllApplications()
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

	domainID, _ := strconv.ParseUint(c.Query("domain_id"), 10, 64)

	var certs []store.Certificate
	var total int64
	var err error

	if domainID > 0 {
		certs, total, err = h.db.ListCertificatesByDomainAndUser(uint(domainID), userID, offset, pageSize)
	} else {
		certs, total, err = h.db.ListCertificatesByUser(userID, offset, pageSize)
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
	response.OK(c, gin.H{"domain": domain, "challenge_value": val})
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
