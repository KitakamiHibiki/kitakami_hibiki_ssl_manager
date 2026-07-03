package acme

import (
	"fmt"
	"sync"
	"time"
)

// ManualProvider implements lego's ChallengeProvider for DNS-01.
// It stores the challenge values and blocks until ResumeChallenge is called.
type ManualProvider struct {
	mu      sync.Mutex
	records map[string]record
}

type record struct {
	keyAuth string
	signal  chan struct{}
}

func NewManualProvider() *ManualProvider {
	return &ManualProvider{records: make(map[string]record)}
}

func (p *ManualProvider) Present(domain, token, keyAuth string) error {
	p.mu.Lock()
	p.records[domain] = record{keyAuth: keyAuth, signal: make(chan struct{})}
	p.mu.Unlock()
	return nil
}

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

// GetKeyAuth returns the key auth value for a domain challenge.
func (p *ManualProvider) GetKeyAuth(domain string) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	r, ok := p.records[domain]
	if !ok {
		return ""
	}
	return r.keyAuth
}

// Signal triggers the challenge for a domain to proceed.
func (p *ManualProvider) Signal(domain string) error {
	p.mu.Lock()
	r, ok := p.records[domain]
	p.mu.Unlock()
	if !ok {
		return fmt.Errorf("no pending challenge for %s", domain)
	}
	close(r.signal)
	return nil
}
