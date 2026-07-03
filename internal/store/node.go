package store

import (
	"time"

	"gorm.io/gorm"
)

type DeployNode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Name      string    `gorm:"not null" json:"name"`
	NodeType  string    `gorm:"default:ssh" json:"node_type"`
	Host      string    `gorm:"not null" json:"host"`
	Port      int       `gorm:"default:22" json:"port"`
	Username  string    `gorm:"not null" json:"username"`
	AuthType  string    `gorm:"default:password" json:"auth_type"`
	Password  string    `gorm:"default:''" json:"-"`
	SSHKey    string    `gorm:"default:''" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (db *DB) CreateNode(n *DeployNode) error {
	return db.Create(n).Error
}

func (db *DB) ListNodeByUser(userID uint) ([]DeployNode, error) {
	var list []DeployNode
	err := db.Where("user_id = ?", userID).Order("id ASC").Find(&list).Error
	return list, err
}

func (db *DB) ListNodeByID(id uint) (*DeployNode, error) {
	var n DeployNode
	err := db.First(&n, id).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (db *DB) UpdateNode(n *DeployNode) error {
	return db.Save(n).Error
}

func (db *DB) DeleteNode(id uint) error {
	return db.Delete(&DeployNode{}, id).Error
}
