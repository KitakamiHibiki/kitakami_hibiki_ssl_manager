package store

import "time"

func (db *DB) GetCertificateByID(id uint) (*Certificate, error) {
	var c Certificate
	err := db.First(&c, id).Error
	if err != nil {
		return nil, err
	}
	var d Domain
	if err := db.First(&d, c.DomainID).Error; err == nil {
		c.Domain = d.Domain
	}
	return &c, nil
}

func (db *DB) CreateCertificate(c *Certificate) error {
	return db.Create(c).Error
}

func (db *DB) UpdateCertificate(c *Certificate) error {
	return db.Model(c).Select("*").Omit("domain_id").Updates(c).Error
}

func (db *DB) GetCertificateByIDAndUser(id, userID uint) (*Certificate, error) {
	var c Certificate
	err := db.First(&c, id).Error
	if err != nil {
		return nil, err
	}
	var d Domain
	err = db.Where("id = ? AND user_id = ?", c.DomainID, userID).First(&d).Error
	if err != nil {
		return nil, err
	}
	c.Domain = d.Domain
	return &c, nil
}

func (db *DB) GetCertificatesByDomainID(domainID uint) ([]Certificate, error) {
	var certs []Certificate
	err := db.Where("domain_id = ?", domainID).Find(&certs).Error
	return certs, err
}

func (db *DB) ListCertificatesByDomainAndUser(domainID, userID uint, offset, limit int) ([]Certificate, int64, error) {
	var certs []Certificate
	var total int64

	if err := db.Table("certificates").
		Joins("JOIN domains ON domains.id = certificates.domain_id").
		Where("domains.user_id = ? AND certificates.domain_id = ?", userID, domainID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Table("certificates").
		Joins("JOIN domains ON domains.id = certificates.domain_id").
		Where("domains.user_id = ? AND certificates.domain_id = ?", userID, domainID).
		Select("certificates.*").
		Order("certificates.id DESC").
		Offset(offset).
		Limit(limit).
		Find(&certs).Error
	if err != nil {
		return nil, 0, err
	}
	for i := range certs {
		var d Domain
		if err := db.First(&d, certs[i].DomainID).Error; err == nil {
			certs[i].Domain = d.Domain
		}
	}
	return certs, total, nil
}

func (db *DB) ListCertificatesByUser(userID uint, offset, limit int) ([]Certificate, int64, error) {
	var certs []Certificate
	var total int64

	if err := db.Table("certificates").
		Joins("JOIN domains ON domains.id = certificates.domain_id").
		Where("domains.user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Table("certificates").
		Joins("JOIN domains ON domains.id = certificates.domain_id").
		Where("domains.user_id = ?", userID).
		Select("certificates.*").
		Order("certificates.id DESC").
		Offset(offset).
		Limit(limit).
		Find(&certs).Error
	if err != nil {
		return nil, 0, err
	}
	for i := range certs {
		var d Domain
		if err := db.First(&d, certs[i].DomainID).Error; err == nil {
			certs[i].Domain = d.Domain
		}
	}
	return certs, total, nil
}

func (db *DB) GetExpiringCertificates(before time.Time) ([]Certificate, error) {
	var certs []Certificate
	err := db.Where("expires_at < ? AND status = ?", before, "issued").Find(&certs).Error
	if err != nil {
		return nil, err
	}
	for i := range certs {
		var d Domain
		if err := db.First(&d, certs[i].DomainID).Error; err == nil {
			certs[i].Domain = d.Domain
		}
	}
	return certs, nil
}

func (db *DB) DeleteCertificate(id uint) error {
	return db.Delete(&Certificate{}, id).Error
}

func (db *DB) DeleteCertificateByIDAndUser(id, userID uint) error {
	var c Certificate
	if err := db.First(&c, id).Error; err != nil {
		return err
	}
	var d Domain
	if err := db.Where("id = ? AND user_id = ?", c.DomainID, userID).First(&d).Error; err != nil {
		return err
	}
	return db.Delete(&Certificate{}, id).Error
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
	for i := range certs {
		var d Domain
		if err := db.First(&d, certs[i].DomainID).Error; err == nil {
			certs[i].Domain = d.Domain
		}
	}
	return certs, total, nil
}

func (db *DB) MarkIncompleteCertificatesAsError() error {
	return db.Model(&Certificate{}).
		Where("status NOT IN ?", []string{"issued", "error"}).
		Updates(map[string]interface{}{
			"status":    "error",
			"error_msg": "由于后端服务终止",
		}).Error
}
