package dns

import (
	"fmt"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/acme"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
)

// StoreAdapter implements ChallengeStore using DB and acme.ManualDNS.
type StoreAdapter struct {
	db *store.DB
}

func NewStoreAdapter(db *store.DB) *StoreAdapter {
	return &StoreAdapter{db: db}
}

func (a *StoreAdapter) GetDomainByHash(hash string) (string, error) {
	for _, app := range acme.GetAllApplications() {
		if app.DomainHash == hash {
			return app.Domain, nil
		}
	}
	return "", fmt.Errorf("hash %s not found", hash)
}

func (a *StoreAdapter) GetChallengeKeyAuth(domain string) string {
	return acme.ManualDNS.GetKeyAuth(domain)
}
