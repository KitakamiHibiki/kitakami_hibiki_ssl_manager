package store

import "time"

type DeployLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CertID     uint      `gorm:"index;not null" json:"cert_id"`
	DomainID   uint      `gorm:"index;not null" json:"domain_id"`
	NodeID     uint      `gorm:"not null" json:"node_id"`
	NodeName   string    `json:"node_name"`
	Status     string    `gorm:"default:pending" json:"status"`
	Detail     string    `json:"detail"`
	ErrorMsg   string    `json:"error_msg"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	CreatedAt  time.Time `json:"created_at"`
}

func (db *DB) CreateDeployLog(dl *DeployLog) error {
	return db.Create(dl).Error
}

func (db *DB) ListDeployLogsByCert(certID uint) ([]DeployLog, error) {
	var logs []DeployLog
	err := db.Where("cert_id = ?", certID).Order("id DESC").Find(&logs).Error
	return logs, err
}

func (db *DB) ListDeployLogsByDomain(domainID uint) ([]DeployLog, error) {
	var logs []DeployLog
	err := db.Where("domain_id = ?", domainID).Order("id DESC").Find(&logs).Error
	return logs, err
}

func (db *DB) ListDeployLogsByDomainPaginated(domainID uint, offset, limit int) ([]DeployLog, int64, error) {
	var logs []DeployLog
	var total int64

	if err := db.Model(&DeployLog{}).Where("domain_id = ?", domainID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Where("domain_id = ?", domainID).Order("id DESC").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func (db *DB) UpdateDeployLog(dl *DeployLog) error {
	return db.Save(dl).Error
}
