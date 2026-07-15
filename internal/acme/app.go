package acme

import (
	"sync"
	"time"
)

// Application tracks an in-progress certificate application.
type Application struct {
	Domain        string    `json:"domain"`
	Email         string    `json:"email"`
	Domains       []string  `json:"domains"`        // all SAN domains (primary first)
	Status        string    `json:"status"`         // pending, issued, error
	ErrorMsg      string    `json:"error_msg"`
	ExpiresAt     time.Time `json:"expires_at"`
	CertID        uint      `json:"cert_id"`
	UserID        uint      `json:"user_id"`
	DomainHash    string    `json:"domain_hash"`
	ChallengeType string    `json:"challenge_type"` // "dns", "http", "cname"
}

var (
	appsMu   sync.Mutex
	apps     = make(map[uint]*Application) // keyed by cert ID
	certApps []*Application                // ordered list for iteration
)

// SetApplication stores an in-progress application.
func SetApplication(a *Application) {
	appsMu.Lock()
	apps[a.CertID] = a
	rebuildList()
	appsMu.Unlock()
}

// RemoveApplication removes an application from the store.
func RemoveApplication(a *Application) {
	appsMu.Lock()
	delete(apps, a.CertID)
	rebuildList()
	appsMu.Unlock()
}

func rebuildList() {
	certApps = make([]*Application, 0, len(apps))
	for _, v := range apps {
		certApps = append(certApps, v)
	}
}

// GetApplication returns the application for a cert ID.
func GetApplication(certID uint) *Application {
	appsMu.Lock()
	defer appsMu.Unlock()
	return apps[certID]
}

// FindApplicationByDomain finds an application that contains the given domain.
func FindApplicationByDomain(domain string) *Application {
	appsMu.Lock()
	defer appsMu.Unlock()
	for _, a := range apps {
		for _, d := range a.Domains {
			if d == domain {
				return a
			}
		}
	}
	return nil
}

// GetAllApplications returns all in-progress applications.
func GetAllApplications() []*Application {
	appsMu.Lock()
	defer appsMu.Unlock()
	return certApps
}
