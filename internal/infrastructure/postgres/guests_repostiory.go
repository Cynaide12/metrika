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

func (r *GuestsRepository) CreateGuests(ctx context.Context, guests *[]domain.Guest) ([]domain.Guest, error) {
	db := getDB(ctx, r.db)

	var mGuests []Guest
	for _, guest := range *guests {
		mGuests = append(mGuests, Guest{
			Fingerprint: guest.Fingerprint,
			DomainID:    guest.DomainID,
		})
	}

	if err := db.Model(&Guest{}).Create(&mGuests).Error; err != nil {
		return nil, err
	}

	var dGuests []domain.Guest
	for _, guest := range mGuests{
		dGuests = append(dGuests, domain.Guest{
			ID: guest.ID,
			DomainID: guest.DomainID,
			Fingerprint: guest.Fingerprint,
		})
	}

	return dGuests, nil
}
