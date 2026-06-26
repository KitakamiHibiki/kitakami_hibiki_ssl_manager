package store

func (db *DB) CreateDomain(d *Domain) error {
	return db.Create(d).Error
}

func (db *DB) ListDomains() ([]Domain, error) {
	var domains []Domain
	err := db.Find(&domains).Error
	return domains, err
}

func (db *DB) GetDomain(id uint) (*Domain, error) {
	var d Domain
	err := db.First(&d, id).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (db *DB) DeleteDomain(id uint) error {
	return db.Delete(&Domain{}, id).Error
}
