package postgres

import (
	"context"
	domain "metrika/internal/domain/analytics"

	"gorm.io/gorm"
)

type DomainRepository struct {
	db *gorm.DB
}

func NewDomainRepository(db *gorm.DB) *DomainRepository {
	return &DomainRepository{db}
}

func (d *DomainRepository) ByURL(ctx context.Context, url string) (*domain.Domain, error) {
	db := getDB(ctx, d.db)

	var mdomain Domain

	if err := db.Model(Domain{}).Where("site_url = ?", url).First(&mdomain).Error; err != nil {
		return nil, err
	}

	return &domain.Domain{ID: mdomain.ID, SiteURL: mdomain.SiteURL}, nil
}
