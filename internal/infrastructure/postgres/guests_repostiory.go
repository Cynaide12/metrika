package postgres

import (
	"context"
	domain "metrika/internal/domain/analytics"

	"gorm.io/gorm"
)

type GuestsRepository struct {
	db *gorm.DB
}

func NewGuestsRepository(db *gorm.DB) *GuestsRepository {
	return &GuestsRepository{db}
}

func (r *GuestsRepository) ByFingerprint(ctx context.Context, fingerprint string) (*domain.Guest, error) {
	db := getDB(ctx, r.db)

	var mGuest Guest

	if err := db.Model(&Guest{}).Where("f_id = ?", fingerprint).First(&mGuest).Error; err != nil {
		return nil, err
	}

	return &domain.Guest{Fingerprint: fingerprint, ID: mGuest.ID, DomainID: mGuest.DomainID}, nil
}

func (r *GuestsRepository) FirstOrCreate(ctx context.Context, fingerprint string, domain_id uint) (*domain.Guest, error) {
	db := getDB(ctx, r.db)

	mGuest := Guest{
		Fingerprint: fingerprint,
		DomainID:    domain_id,
	}

	if err := db.Model(&Guest{}).Where("f_id = ?", fingerprint).FirstOrCreate(&mGuest).Error; err != nil {
		return nil, err
	}

	return &domain.Guest{DomainID: domain_id, Fingerprint: fingerprint, ID: mGuest.ID}, nil
}
