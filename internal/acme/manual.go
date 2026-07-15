package acme

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// ManualProvider implements lego's ChallengeProvider for DNS-01 and HTTP-01.
// It stores challenge values and blocks until Signal is called for each domain.
type ManualProvider struct {
	mu       sync.Mutex
	records  map[string]record     // DNS-01 records (keyed by domain)
	httpRecs map[string]httpRecord // HTTP-01 records (keyed by domain)
}

type record struct {
	keyAuth  string
	signal   chan struct{}
	signaled bool
}

type httpRecord struct {
	token    string
	keyAuth  string
	signal   chan struct{}
	signaled bool
}

func NewManualProvider() *ManualProvider {
	return &ManualProvider{
		records:  make(map[string]record),
		httpRecs: make(map[string]httpRecord),
	}
}

// Present stores the challenge for DNS-01.
func (p *ManualProvider) Present(domain, token, keyAuth string) error {
	p.mu.Lock()
	p.records[domain] = record{keyAuth: keyAuth, signal: make(chan struct{})}
	p.mu.Unlock()
	return nil
}

// CleanUp removes a DNS-01 challenge record.
func (p *ManualProvider) CleanUp(domain, token, keyAuth string) error {
	p.mu.Lock()
	delete(p.records, domain)
	p.mu.Unlock()
	return nil
}

// Timeout is used by lego to know how long to wait.
func (p *ManualProvider) Timeout() (time.Duration, time.Duration) {
	return 30 * time.Second, 2 * time.Second
}

// GetKeyAuth returns the DNS-01 TXT record value (base64url-encoded SHA256 of key auth).
func (p *ManualProvider) GetKeyAuth(domain string) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	r, ok := p.records[domain]
	if !ok {
		return ""
	}
	hash := sha256.Sum256([]byte(r.keyAuth))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// GetHTTPChallenge returns the HTTP-01 challenge token and keyAuth for a domain.
func (p *ManualProvider) GetHTTPChallenge(domain string) (token, keyAuth string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	r, ok := p.httpRecs[domain]
	if !ok {
		return "", ""
	}
	return r.token, r.keyAuth
}

// Signal triggers the DNS-01 challenge for a domain to proceed.
func (p *ManualProvider) Signal(domain string) error {
	p.mu.Lock()
	r, ok := p.records[domain]
	if !ok {
		p.mu.Unlock()
		return fmt.Errorf("no pending challenge for %s", domain)
	}
	if !r.signaled {
		r.signaled = true
		p.records[domain] = r
		close(r.signal)
	}
	p.mu.Unlock()
	return nil
}

// HasDomain returns true if a DNS-01 challenge exists for the given domain.
func (p *ManualProvider) HasDomain(domain string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.records[domain]
	return ok
}

// IsSignaled returns true if the DNS-01 challenge for the domain has been signaled.
func (p *ManualProvider) IsSignaled(domain string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	r, ok := p.records[domain]
	return ok && r.signaled
}

// HTTPPresent stores the challenge for HTTP-01.
func (p *ManualProvider) HTTPPresent(domain, token, keyAuth string) error {
	p.mu.Lock()
	p.httpRecs[domain] = httpRecord{token: token, keyAuth: keyAuth, signal: make(chan struct{})}
	p.mu.Unlock()
	return nil
}

// HTTPCleanUp removes an HTTP-01 challenge record.
func (p *ManualProvider) HTTPCleanUp(domain, token, keyAuth string) error {
	p.mu.Lock()
	delete(p.httpRecs, domain)
	p.mu.Unlock()
	return nil
}

// HTTPSignal triggers the HTTP-01 challenge for a domain to proceed.
func (p *ManualProvider) HTTPSignal(domain string) error {
	p.mu.Lock()
	r, ok := p.httpRecs[domain]
	if !ok {
		p.mu.Unlock()
		return fmt.Errorf("no pending HTTP challenge for %s", domain)
	}
	if !r.signaled {
		r.signaled = true
		p.httpRecs[domain] = r
		close(r.signal)
	}
	p.mu.Unlock()
	return nil
}

// HasHTTPDomain returns true if an HTTP-01 challenge exists for the given domain.
func (p *ManualProvider) HasHTTPDomain(domain string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.httpRecs[domain]
	return ok
}

// IsHTTPSignaled returns true if the HTTP-01 challenge for the domain has been signaled.
func (p *ManualProvider) IsHTTPSignaled(domain string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	r, ok := p.httpRecs[domain]
	return ok && r.signaled
}

// GetKeyAuthByToken looks up the keyAuth for an HTTP-01 challenge by token.
func (p *ManualProvider) GetKeyAuthByToken(token string) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, r := range p.httpRecs {
		if r.token == token {
			return r.keyAuth
		}
	}
	return ""
}

// HTTPManualProvider wraps ManualProvider to implement challenge.Provider for HTTP-01.
// It delegates Present/CleanUp to the HTTP-specific methods on ManualProvider.
type HTTPManualProvider struct {
	parent *ManualProvider
}

// NewHTTPManualProvider creates a new HTTP-01 challenge provider.
func NewHTTPManualProvider(parent *ManualProvider) *HTTPManualProvider {
	return &HTTPManualProvider{parent: parent}
}

func (h *HTTPManualProvider) Present(domain, token, keyAuth string) error {
	return h.parent.HTTPPresent(domain, token, keyAuth)
}

func (h *HTTPManualProvider) CleanUp(domain, token, keyAuth string) error {
	return h.parent.HTTPCleanUp(domain, token, keyAuth)
}
