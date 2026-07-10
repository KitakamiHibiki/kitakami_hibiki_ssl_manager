package acme

import (
	"sync"
	"time"
)

// Application tracks an in-progress certificate application.
type Application struct {
	Domain    string    `json:"domain"`
	DomainID  uint      `json:"domain_id"`
	Domains   []string  `json:"domains"`   // all SAN domains (primary first)
	Status    string    `json:"status"`    // pending, issued, error
	ErrorMsg  string    `json:"error_msg"`
	ExpiresAt time.Time `json:"expires_at"`
	CertID    uint      `json:"cert_id"`
}

var (
	appsMu   sync.Mutex
	apps     = make(map[string]*Application) // keyed by primary domain
)

// SetApplication stores an in-progress application.
func SetApplication(domain string, a *Application) {
	appsMu.Lock()
	apps[domain] = a
	appsMu.Unlock()
}

// GetApplication returns the application for a domain.
func GetApplication(domain string) *Application {
	appsMu.Lock()
	defer appsMu.Unlock()
	return apps[domain]
}

// GetAllApplications returns all in-progress applications.
func GetAllApplications() []*Application {
	appsMu.Lock()
	defer appsMu.Unlock()
	list := make([]*Application, 0, len(apps))
	for _, a := range apps {
		list = append(list, a)
	}
	return list
}
