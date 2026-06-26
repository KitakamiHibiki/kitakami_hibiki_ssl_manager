package scheduler

import (
	"log"
	"time"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/config"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
	db   *store.DB
	cfg  *config.Config
}

// RenewFunc is the callback for certificate renewal, accepting (domain, email, certDir).
type RenewFunc func(domain, email, certDir string) error

func New(cfg *config.Config, db *store.DB) *Scheduler {
	return &Scheduler{
		cron: cron.New(),
		db:   db,
		cfg:  cfg,
	}
}

func (s *Scheduler) Start(renewFn RenewFunc) error {
	_, err := s.cron.AddFunc(s.cfg.Sched.CheckInterval, func() {
		s.checkAndRenew(renewFn)
	})
	if err != nil {
		return err
	}
	s.cron.Start()
	log.Println("[scheduler] started, check interval:", s.cfg.Sched.CheckInterval)
	return nil
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}

func (s *Scheduler) checkAndRenew(renewFn RenewFunc) {
	before := time.Now().AddDate(0, 0, s.cfg.Sched.RenewBeforeDays)
	certs, err := s.db.GetExpiringCertificates(before)
	if err != nil {
		log.Println("[scheduler] query expiring certs error:", err)
		return
	}

	for _, cert := range certs {
		log.Printf("[scheduler] renewing cert for domain: %s (expires: %s)", cert.Domain, cert.ExpiresAt)
		if err := renewFn(cert.Domain, "", ""); err != nil {
			log.Printf("[scheduler] renew failed for %s: %v", cert.Domain, err)
			continue
		}
		cert.ExpiresAt = time.Now().AddDate(0, 3, 0)
		cert.Status = "issued"
		if err := s.db.UpdateCertificate(&cert); err != nil {
			log.Printf("[scheduler] update cert record error: %v", err)
		}
	}
}
