package store

import (
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Domain struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Domain    string    `gorm:"uniqueIndex:idx_domain_active,where:deleted_at IS NULL;not null" json:"domain"`
	Email     string    `gorm:"not null" json:"email"`
	Challenge string    `gorm:"default:dns" json:"challenge"` // dns only
	DeployEnabled bool   `gorm:"default:false" json:"deploy_enabled"`
	DeployNodeID  uint   `gorm:"default:0" json:"deploy_node_id"`
	DeployType    string `gorm:"default:nginx" json:"deploy_type"`
	CertName      string `gorm:"default:fullchain.pem" json:"cert_name"`
	CertPath      string `gorm:"default:/etc/nginx/certs" json:"cert_path"`
	KeyName       string `gorm:"default:privkey.key" json:"key_name"`
	KeyPath       string `gorm:"default:/etc/nginx/certs" json:"key_path"`
	AutoRenew     bool   `gorm:"default:false" json:"auto_renew"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Certificate struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DomainID  uint      `gorm:"index;not null" json:"domain_id"`
	Domain    string    `gorm:"-" json:"domain"`
	Domains   string    `gorm:"default:'[]'" json:"domains"` // JSON array of all SAN domains
	Status    string    `gorm:"default:pending" json:"status"`
	ErrorMsg  string    `gorm:"default:''" json:"error_msg"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type DB struct {
	*gorm.DB
}

func InitDB(dsn string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Domain{}, &Certificate{}, &User{}, &SystemConfig{}, &DeployNode{}); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// AfterInit runs tasks that must follow InitDB: migrations and seed data.
func (db *DB) AfterInit() error {
	if err := db.RunMigrations(); err != nil {
		return err
	}
	db.SeedAdmin()
	if err := db.SeedSystemConfig(); err != nil {
		return err
	}
	return nil
}
