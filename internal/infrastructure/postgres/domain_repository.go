package postgres

import (
	"context"
	"errors"
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDomainNotFound
		}
		return nil, err
	}

	return &domain.Domain{ID: mdomain.ID, SiteURL: mdomain.SiteURL}, nil
}

func (d *DomainRepository) AddDomain(ctx context.Context, site_url string) (*domain.Domain, error) {
	db := getDB(ctx, d.db)

	dom := Domain{
		SiteURL: site_url,
	}

	if err := db.Model(&Domain{}).Create(&dom).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDomainAlreadyExists
		}
	}

	return &domain.Domain{SiteURL: site_url, ID: dom.ID}, nil
}

func (d *DomainRepository) GetDomainGuests(ctx context.Context, domainId uint) (*[]domain.Guest, error) {
	db := getDB(ctx, d.db)

	var mGuests []Guest

	if err := db.Model(&Guest{}).Where("domain_id = ?", domainId).Find(&mGuests).Error; err != nil {
		return nil, err
	}

	var guests []domain.Guest

	for _, guest := range mGuests {
		guests = append(guests, domain.Guest{
			ID:          guest.ID,
			DomainID:    guest.DomainID,
			Fingerprint: guest.Fingerprint,
		})
	}

	return &guests, nil
}

func (d *DomainRepository) GetDomainGuestsByFingerprints(ctx context.Context, domainId uint, fingerprints []string) (*[]domain.Guest, error) {
	db := getDB(ctx, d.db)

	var mGuests []Guest

	if err := db.Model(&Guest{}).Where("domain_id = ? AND f_id IN ?", domainId, fingerprints).Find(&mGuests).Error; err != nil {
		return nil, err
	}

	var guests []domain.Guest

	for _, guest := range mGuests {
		guests = append(guests, domain.Guest{
			ID:          guest.ID,
			DomainID:    guest.DomainID,
			Fingerprint: guest.Fingerprint,
		})
	}

	return &guests, nil
}

func (d *DomainRepository) GetCountDomainGuests(ctx context.Context, domain_id uint) (int64, error) {
	db := getDB(ctx, d.db)

	var count int64

	if err := db.Model(&Domain{}).Where("id = ?", domain_id).Count(&count).Error; err != nil {
		if errors.Is(err, domain.ErrDomainNotFound) {
			return 0, domain.ErrDomainNotFound
		}
		return 0, err
	}

	return count, nil
}
