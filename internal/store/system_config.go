package store

import "time"

// SystemConfig holds runtime configuration persisted in the database.
// Only bootstrap config (server.port, storage.driver, storage.dsn) stays in config.yaml.
type SystemConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	JWTSecret string    `gorm:"default:change-me-in-production" json:"jwt_secret"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (db *DB) GetSystemConfig() (*SystemConfig, error) {
	var cfg SystemConfig
	err := db.First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (db *DB) UpdateSystemConfig(cfg *SystemConfig) error {
	return db.Save(cfg).Error
}

func (db *DB) SeedSystemConfig() error {
	var count int64
	db.Model(&SystemConfig{}).Count(&count)
	if count == 0 {
		return db.Create(&SystemConfig{}).Error
	}
	return nil
}
