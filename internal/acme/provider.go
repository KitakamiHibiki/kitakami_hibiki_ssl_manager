package acme

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type HTTPProvider struct {
	mu    sync.RWMutex
	token map[string]string
}

func NewHTTPProvider() *HTTPProvider {
	return &HTTPProvider{token: make(map[string]string)}
}

func (p *HTTPProvider) Present(domain, token, keyAuth string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.token[token] = keyAuth
	return nil
}

func (p *HTTPProvider) CleanUp(domain, token, keyAuth string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.token, token)
	return nil
}

func (p *HTTPProvider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const prefix = "/.well-known/acme-challenge/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		http.NotFound(w, r)
		return
	}

	token := strings.TrimPrefix(r.URL.Path, prefix)
	if token == "" || strings.Contains(token, "/") {
		http.NotFound(w, r)
		return
	}

	p.mu.RLock()
	keyAuth, ok := p.token[token]
	p.mu.RUnlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, keyAuth)
}
