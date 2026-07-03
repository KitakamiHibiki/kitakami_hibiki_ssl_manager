package store

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex:idx_email_active,where:deleted_at IS NULL;not null" json:"email"`
	Username  string    `gorm:"not null" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	Role      string         `gorm:"default:user" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (db *DB) CreateUser(u *User) error {
	return db.Create(u).Error
}

func (db *DB) FindUserByUsername(username string) (*User, error) {
	var u User
	err := db.Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (db *DB) FindUserByEmail(email string) (*User, error) {
	var u User
	err := db.Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (db *DB) GetUserByID(id uint) (*User, error) {
	var u User
	err := db.First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (db *DB) ListUsers() ([]User, error) {
	var users []User
	err := db.Find(&users).Error
	return users, err
}

func (db *DB) UpdateUserRole(u *User) error {
	return db.Model(u).Update("role", u.Role).Error
}

func (db *DB) DeleteUser(id uint) error {
	return db.Delete(&User{}, id).Error
}

func (db *DB) SeedAdmin() {
	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		db.Create(&User{
			Email:    "admin@khssl.local",
			Username: "admin",
		Password: "$2a$10$KlbyfqMxDkB.XW.HoJBZw.pXQaHLVUdd25Ky7fgZSrV6KdTW7vehy", // admin123
			Role:     "admin",
		})
	}
}
