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
// Append new migrations at the end as the project evolves.
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
				// Drop old full unique indexes
				if err := tx.Migrator().DropIndex(&Domain{}, "idx_domains_domain"); err != nil {
					// Index might not exist, ignore
				}
				if err := tx.Migrator().DropIndex(&User{}, "idx_users_email"); err != nil {
					// Index might not exist, ignore
				}
				// Create new partial unique indexes
				if err := tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_domain_active ON domains(domain) WHERE deleted_at IS NULL").Error; err != nil {
					return err
				}
				if err := tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_email_active ON users(email) WHERE deleted_at IS NULL").Error; err != nil {
					return err
				}
				return nil
			},
		},
	}
}

// RunMigrations applies all pending migrations in order.
// Each migration runs inside a transaction so a failure rolls back cleanly.
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
