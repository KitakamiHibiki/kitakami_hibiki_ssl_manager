package store

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Domain struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Domain    string    `gorm:"uniqueIndex;not null" json:"domain"`
	Email     string    `gorm:"not null" json:"email"`
	Challenge string    `gorm:"default:http" json:"challenge"` // http | dns
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Certificate struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DomainID  uint      `gorm:"index;not null" json:"domain_id"`
	Domain    string    `gorm:"-" json:"domain"`
	Status    string    `gorm:"default:pending" json:"status"` // pending | issued | expired | error
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
	if err := db.AutoMigrate(&Domain{}, &Certificate{}); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}
