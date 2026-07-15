package store

import "time"

func (db *DB) GetCertificateByID(id uint) (*Certificate, error) {
	var c Certificate
	err := db.First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (db *DB) CreateCertificate(c *Certificate) error {
	return db.Create(c).Error
}

func (db *DB) UpdateCertificate(id uint, updates map[string]interface{}) error {
	return db.Model(&Certificate{}).Where("id = ?", id).Updates(updates).Error
}

func (db *DB) GetCertificateByIDAndUser(id, userID uint) (*Certificate, error) {
	var c Certificate
	err := db.Where("id = ? AND user_id = ?", id, userID).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (db *DB) DeleteCertificate(id uint) error {
	return db.Delete(&Certificate{}, id).Error
}

func (db *DB) DeleteCertificateByIDAndUser(id, userID uint) error {
	return db.Where("id = ? AND user_id = ?", id, userID).Delete(&Certificate{}).Error
}

func (db *DB) ListCertificatesByUser(userID uint, offset, limit int) ([]Certificate, int64, error) {
	var certs []Certificate
	var total int64

	if err := db.Model(&Certificate{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Where("user_id = ?", userID).Order("id DESC").Offset(offset).Limit(limit).Find(&certs).Error
	if err != nil {
		return nil, 0, err
	}
	return certs, total, nil
}

func (db *DB) ListAllCertificates(offset, limit int) ([]Certificate, int64, error) {
	var certs []Certificate
	var total int64

	if err := db.Model(&Certificate{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Order("id DESC").Offset(offset).Limit(limit).Find(&certs).Error
	if err != nil {
		return nil, 0, err
	}
	return certs, total, nil
}

func (db *DB) GetExpiringCertificates(before time.Time) ([]Certificate, error) {
	var certs []Certificate
	err := db.Where("expires_at < ? AND status = ?", before, "issued").Find(&certs).Error
	if err != nil {
		return nil, err
	}
	return certs, nil
}

func (db *DB) MarkIncompleteCertificatesAsError() error {
	return db.Model(&Certificate{}).
		Where("status NOT IN ?", []string{"issued", "error"}).
		Updates(map[string]interface{}{
			"status":    "error",
			"error_msg": "由于后端服务终止",
		}).Error
}
