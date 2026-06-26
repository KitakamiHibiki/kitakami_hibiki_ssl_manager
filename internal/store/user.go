package store

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	Role      string    `gorm:"default:user" json:"role"`
	CreatedAt time.Time `json:"created_at"`
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

func (db *DB) SeedAdmin() {
	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		db.Create(&User{
			Username: "admin",
			Password: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // admin123
			Role:     "admin",
		})
	}
}
