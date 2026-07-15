package handler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/acme"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/config"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/dns"
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
	dnsAddr string
}

func NewCertHandler(db *store.DB, cfg *config.Config, certDir string, dnsAddr string) *CertHandler {
	return &CertHandler{db: db, cfg: cfg, certDir: certDir, dnsAddr: dnsAddr}
}

func (h *CertHandler) Apply(c *gin.Context) {
	var req struct {
		Domain        string   `json:"domain"`
		Email         string   `json:"email"`
		ExtraDomains  []string `json:"extra_domains"`
		ChallengeType string   `json:"challenge_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Domain == "" {
		response.Error(c, 400, "domain is required")
		return
	}
	if req.ChallengeType == "" {
		req.ChallengeType = "http"
	}

	userID := c.GetUint("user_id")
	email := req.Email
	if email == "" {
		email = c.GetString("email")
	}

	allDomains := []string{req.Domain}
	allDomains = append(allDomains, req.ExtraDomains...)

	root := extractRootDomain(req.Domain)
	for _, extra := range req.ExtraDomains {
		if extractRootDomain(extra) != root {
			response.Error(c, 400, "extra domain '"+extra+"' does not share the same root domain as '"+req.Domain+"'")
			return
		}
	}

	seen := map[string]bool{}
	unique := allDomains[:0]
	for _, name := range allDomains {
		if !seen[name] {
			seen[name] = true
			unique = append(unique, name)
		}
	}
	allDomains = unique

	app := &acme.Application{
		Domain:        req.Domain,
		Domains:       allDomains,
		Status:        "verifying",
		UserID:        userID,
		Email:         email,
		DomainHash:    md5Hex(req.Domain),
		ChallengeType: req.ChallengeType,
	}
	acme.SetApplication(app)

	response.OK(c, gin.H{"domain": req.Domain, "domains": allDomains, "domain_hash": app.DomainHash, "status": "verifying"})
}

func (h *CertHandler) VerifyHTTPProxy(c *gin.Context) {
	var req struct {
		Domain     string `json:"domain"`
		DomainHash string `json:"domain_hash"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Domain == "" {
		response.Error(c, 400, "domain is required")
		return
	}

	verifyToken := "ssl-verify"
	if req.DomainHash != "" {
		verifyToken = "ssl-verify-" + req.DomainHash
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	httpReq, _ := http.NewRequestWithContext(ctx, "GET",
		"http://"+req.Domain+"/.well-known/acme-challenge/"+verifyToken, nil)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		response.Error(c, 400, "代理验证失败：无法连接到 "+req.Domain+"，请确认代理配置已生效")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ssl-verify-ok" {
		response.Error(c, 400, "代理验证失败：响应不匹配，请确认已将 "+req.Domain+" 的 /.well-known/acme-challenge/ 路径代理到本服务器")
		return
	}

	response.OK(c, gin.H{"domain": req.Domain, "status": "ok"})
}

func (h *CertHandler) VerifyDNS(c *gin.Context) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Domain == "" {
		response.Error(c, 400, "domain is required")
		return
	}

	app := acme.FindApplicationByDomain(req.Domain)
	if app == nil {
		response.Error(c, 404, "no pending application")
		return
	}

	if app.ChallengeType == "http" {
		verifyToken := "ssl-verify-" + app.DomainHash
		proxyCtx, proxyCancel := context.WithTimeout(context.Background(), 8*time.Second)
		proxyReq, _ := http.NewRequestWithContext(proxyCtx, "GET",
			"http://"+req.Domain+"/.well-known/acme-challenge/"+verifyToken, nil)
		proxyResp, proxyErr := http.DefaultClient.Do(proxyReq)
		proxyCancel()
		if proxyErr != nil {
			response.Error(c, 400, "无法连接到 "+req.Domain+"，请确认代理配置已生效")
			return
		}
		if proxyResp != nil {
			body, _ := io.ReadAll(proxyResp.Body)
			proxyResp.Body.Close()
			if string(body) != "ssl-verify-ok" {
				response.Error(c, 400, "代理验证失败：请确认已将 "+req.Domain+" 的 /.well-known/acme-challenge/ 代理到本服务器")
				return
			}
		}

		acme.ManualDNS.HTTPSignal(req.Domain)
	} else if app.ChallengeType == "cname" {
		host := "localhost"
		if u, err := url.Parse(h.cfg.Server.ProxyURL); err == nil && u.Hostname() != "" {
			host = u.Hostname()
		}
		queryName := fmt.Sprintf("%s.challenge.%s", app.DomainHash, host)
		names, err := dns.ResolveTXT(h.dnsAddr, queryName)
		if err != nil {
			response.Error(c, 400, "CNAME 验证失败，无法查询 DNS 记录："+err.Error())
			return
		}
		if len(names) == 0 {
			response.Error(c, 400, "CNAME 记录未生效，请确认已添加 CNAME 记录")
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
			response.Error(c, 400, "CNAME 记录值不匹配，请确认已更新")
			return
		}

		acme.ManualDNS.Signal(req.Domain)
	}

	allDone := true
	var pendingList []string
	for _, d := range app.Domains {
		if app.ChallengeType == "http" {
			if !acme.ManualDNS.IsHTTPSignaled(d) {
				allDone = false
				pendingList = append(pendingList, d)
			}
		} else {
			if !acme.ManualDNS.IsSignaled(d) {
				allDone = false
				pendingList = append(pendingList, d)
			}
		}
	}

	if allDone {
		domainsJSON, _ := json.Marshal(app.Domains)
		certRecord := &store.Certificate{
			UserID:  app.UserID,
			Domain:  app.Domain,
			Domains: string(domainsJSON),
			Email:   app.Email,
			Status:  "issuing",
		}
		if err := h.db.CreateCertificate(certRecord); err != nil {
			response.Error(c, 500, err.Error())
			return
		}
		acme.RemoveApplication(app)
		app.CertID = certRecord.ID
		acme.SetApplication(app)
		app.Status = "issuing"
		go h.startObtain(app)
		response.OK(c, gin.H{"status": "issuing", "cert_id": app.CertID, "all_verified": true})
	} else {
		response.OK(c, gin.H{"status": "verifying", "cert_id": app.CertID, "all_verified": false, "pending": pendingList})
	}
}

func (h *CertHandler) startObtain(app *acme.Application) {
	client, err := acme.NewClient(app.Email, h.cfg.ACME.Directory, h.certDir)
	if err != nil {
		log.Printf("[acme] new client error: %v", err)
		app.Status = "error"
		app.ErrorMsg = acme.CleanError(err)
		h.db.UpdateCertificate(app.CertID, map[string]interface{}{"status": "error", "error_msg": acme.CleanError(err)})
		return
	}

	if app.ChallengeType == "http" {
		client.SetHTTPProvider()
	} else {
		client.SetDNSProvider(h.cfg.ACME.RecursiveNameservers)
	}

	result, err := client.Obtain(app.Domains)
	if err != nil {
		log.Printf("[acme] obtain error for %s: %v", app.Domain, err)
		app.Status = "error"
		app.ErrorMsg = acme.CleanError(err)
		h.db.UpdateCertificate(app.CertID, map[string]interface{}{"status": "error", "error_msg": acme.CleanError(err)})
		return
	}

	certPEM := []byte(result.Certificate)
	expiry, _ := acme.ParseExpiry(certPEM)

	certPath := h.certDir + "/" + app.Domain
	os.MkdirAll(certPath, 0755)
	os.WriteFile(certPath+"/fullchain.pem", certPEM, 0644)
	os.WriteFile(certPath+"/privkey.pem", result.PrivateKey, 0600)

	h.db.UpdateCertificate(app.CertID, map[string]interface{}{
		"status":     "issued",
		"issued_at":  time.Now(),
		"expires_at": expiry,
	})
	app.Status = "issued"
	app.ExpiresAt = expiry
}

func getAllApps() []*acme.Application {
	return acme.GetAllApplications()
}

func (h *CertHandler) Status(c *gin.Context) {
	certID, err := strconv.ParseUint(c.Query("cert_id"), 10, 64)
	if err != nil || certID == 0 {
		response.Error(c, 400, "cert_id is required")
		return
	}
	app := acme.GetApplication(uint(certID))
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

	token, keyAuth := acme.ManualDNS.GetHTTPChallenge(domain)
	if token != "" {
		response.OK(c, gin.H{
			"domain":          domain,
			"challenge_type":  "http",
			"challenge_token": token,
			"challenge_value": keyAuth,
		})
		return
	}

	val := acme.ManualDNS.GetKeyAuth(domain)
	if val == "" {
		response.Error(c, 404, "no pending challenge for this domain")
		return
	}
	response.OK(c, gin.H{
		"domain":          domain,
		"challenge_type":  "dns",
		"challenge_value": val,
	})
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

func md5Hex(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}
