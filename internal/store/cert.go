package store

import "time"

func (db *DB) CreateCertificate(c *Certificate) error {
	return db.Create(c).Error
}

func (db *DB) UpdateCertificate(c *Certificate) error {
	return db.Save(c).Error
}

func (db *DB) ListCertificates() ([]Certificate, error) {
	var certs []Certificate
	err := db.Preload("Domain").Find(&certs).Error
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

func (db *DB) GetCertificate(id uint) (*Certificate, error) {
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

func (db *DB) GetCertificatesByDomainID(domainID uint) ([]Certificate, error) {
	var certs []Certificate
	err := db.Where("domain_id = ?", domainID).Find(&certs).Error
	return certs, err
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
