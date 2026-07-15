package store

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// SchemaMigration tracks which migrations have been applied to the database.
type SchemaMigration struct {
	Version     string    `gorm:"primaryKey;size:50" json:"version"`
	Description string    `gorm:"not null;size:255" json:"description"`
	AppliedAt   time.Time `gorm:"autoCreateTime" json:"applied_at"`
}

// Migration defines a single migration step.
type Migration struct {
	Version     string
	Description string
	Up          func(tx *gorm.DB) error
}

// allMigrations returns the ordered list of all migrations.
func allMigrations() []Migration {
	return []Migration{
		{
			Version:     "2026-07-01-001",
			Description: "Record initial schema (User, Domain, Certificate, SystemConfig)",
			Up:          func(tx *gorm.DB) error { return nil },
		},
		{
			Version:     "2026-07-01-002",
			Description: "Seed SystemConfig default row",
			Up: func(tx *gorm.DB) error {
				var count int64
				tx.Model(&SystemConfig{}).Count(&count)
				if count == 0 {
					return tx.Create(&SystemConfig{}).Error
				}
				return nil
			},
		},
		{
			Version:     "2026-07-02-001",
			Description: "Rebuild unique indexes as partial (WHERE deleted_at IS NULL) for soft delete",
			Up: func(tx *gorm.DB) error {
				tx.Exec("DROP INDEX IF EXISTS idx_users_email")
				if err := tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_email_active ON users(email) WHERE deleted_at IS NULL").Error; err != nil {
					return err
				}
				return nil
			},
		},
		{
			Version:     "2026-07-03-001",
			Description: "Create deploy_logs table",
			Up: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&DeployLog{})
			},
		},
		{
			Version:     "2026-07-03-002",
			Description: "Add cert/key name and path fields to domains",
			Up:          func(tx *gorm.DB) error { return nil },
		},
		{
			Version:     "2026-07-03-003",
			Description: "Add auto_renew field to domains",
			Up:          func(tx *gorm.DB) error { return nil },
		},
		{
			Version:     "2026-07-03-004",
			Description: "Add detail field to deploy_logs",
			Up: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&DeployLog{})
			},
		},
		{
			Version:     "2026-07-10-001",
			Description: "Add domains column to certificates for multi-domain SAN support",
			Up: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&Certificate{})
			},
		},
		{
			Version:     "2026-07-10-002",
			Description: "Merge domains into certificates, make certificates the primary unit",
			Up: func(tx *gorm.DB) error {
				// Check if this is an upgrade (domains table exists) or fresh install
				var hasDomains bool
				tx.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='domains'").Scan(&hasDomains)
				if !hasDomains {
					// Fresh install: Certificate already has all columns from InitDB AutoMigrate.
					// Remove old domain_id if it still exists (from very old schema).
					if tx.Migrator().HasColumn(&Certificate{}, "domain_id") {
						tx.Migrator().DropColumn(&Certificate{}, "domain_id")
					}
					return nil
				}
				// Upgrade path: add columns, copy data, drop domains table
				cols := []struct{ name, def string }{
					{"user_id", "0"},
					{"email", "''"},
					{"deploy_enabled", "0"},
					{"deploy_node_id", "0"},
					{"deploy_type", "'nginx'"},
					{"cert_name", "'fullchain.pem'"},
					{"cert_path", "'/etc/nginx/certs'"},
					{"key_name", "'privkey.key'"},
					{"key_path", "'/etc/nginx/certs'"},
					{"auto_renew", "0"},
				}
				for _, c := range cols {
					tx.Exec("ALTER TABLE certificates ADD COLUMN " + c.name + " INTEGER DEFAULT " + c.def)
				}
				tx.Exec("ALTER TABLE certificates ADD COLUMN domain TEXT DEFAULT ''")
				if err := tx.Exec(`UPDATE certificates SET 
					user_id = (SELECT user_id FROM domains WHERE domains.id = certificates.domain_id),
					email = COALESCE((SELECT email FROM domains WHERE domains.id = certificates.domain_id), ''),
					domain = COALESCE((SELECT domain FROM domains WHERE domains.id = certificates.domain_id), ''),
					deploy_enabled = COALESCE((SELECT deploy_enabled FROM domains WHERE domains.id = certificates.domain_id), 0),
					deploy_node_id = COALESCE((SELECT deploy_node_id FROM domains WHERE domains.id = certificates.domain_id), 0),
					deploy_type = COALESCE((SELECT deploy_type FROM domains WHERE domains.id = certificates.domain_id), 'nginx'),
					cert_name = COALESCE((SELECT cert_name FROM domains WHERE domains.id = certificates.domain_id), 'fullchain.pem'),
					cert_path = COALESCE((SELECT cert_path FROM domains WHERE domains.id = certificates.domain_id), '/etc/nginx/certs'),
					key_name = COALESCE((SELECT key_name FROM domains WHERE domains.id = certificates.domain_id), 'privkey.key'),
					key_path = COALESCE((SELECT key_path FROM domains WHERE domains.id = certificates.domain_id), '/etc/nginx/certs'),
					auto_renew = COALESCE((SELECT auto_renew FROM domains WHERE domains.id = certificates.domain_id), 0)
				`).Error; err != nil {
					return err
				}
				tx.Migrator().DropColumn(&Certificate{}, "domain_id")
				// Drop domains table
				if err := tx.Migrator().DropTable("domains"); err != nil {
					return err
				}
				return nil
			},
		},
	}
}

// RunMigrations applies all pending migrations in order.
func (db *DB) RunMigrations() error {
	if err := db.AutoMigrate(&SchemaMigration{}); err != nil {
		return fmt.Errorf("create migrations tracking table: %w", err)
	}

	var applied []string
	db.Model(&SchemaMigration{}).Pluck("version", &applied)
	appliedSet := make(map[string]bool, len(applied))
	for _, v := range applied {
		appliedSet[v] = true
	}

	for _, m := range allMigrations() {
		if appliedSet[m.Version] {
			continue
		}
		log.Printf("[migrate] running %s: %s", m.Version, m.Description)

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := m.Up(tx); err != nil {
				return err
			}
			return tx.Create(&SchemaMigration{
				Version:     m.Version,
				Description: m.Description,
			}).Error
		})

		if err != nil {
			return fmt.Errorf("migration %s failed: %w", m.Version, err)
		}
		log.Printf("[migrate] applied %s", m.Version)
	}

	return nil
}

// AppliedMigrations returns the list of successfully applied migrations.
func (db *DB) AppliedMigrations() ([]SchemaMigration, error) {
	var list []SchemaMigration
	err := db.Order("version ASC").Find(&list).Error
	return list, err
}
