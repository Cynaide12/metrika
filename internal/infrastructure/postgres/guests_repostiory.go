package postgres

import (
	"context"
	"errors"
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
	for _, guest := range mGuests {
		dGuests = append(dGuests, domain.Guest{
			ID:          guest.ID,
			DomainID:    guest.DomainID,
			Fingerprint: guest.Fingerprint,
		})
	}

	return dGuests, nil
}

// TODO:доделать
func (r *GuestsRepository) Find(ctx context.Context, opts domain.FindGuestsOptions) (*[]domain.Guest, error) {
	db := getDB(ctx, r.db)

	var mGuests *[]Guest

	query := db.Table("guests g").
		Select(`
	g.id,
	g.domain_id, 
	g.f_id,
	MIN(gs.created_at) AS first_visit,
	MAX(gs.created_at) AS last_visit,
	COUNT(gs.id) AS sessions_count,
	EXISTS(SELECT 1 FROM guest_sessions as
	WHERE as.guest_id=g.id AND as.active=true AND as.end_time IS NULL) as online
	`).Joins(`
	LEFT JOIN guest_sessions gs ON g.id=gs.guest_id
	`).Where("g.domain_id=?", opts.DomainID)

	if opts.StartDate != nil && opts.EndDate != nil {
		query = query.Where("EXISTS (SELECT 1 FROM guest_sessions gs2 WHERE gs2.created_at >= ? AND gs2.created_at <= ?)", opts.StartDate, opts.EndDate)
	} else if opts.StartDate != nil {
		query = query.Where("EXISTS (SELECT 1 FROM guest_sessions gs2 WHERE gs2.created_at >= ?)", opts.StartDate)
	} else if opts.EndDate != nil {
		query = query.Where("EXISTS (SELECT 1 FROM guest_sessions gs2 WHERE gs2.created_at <= ?)", opts.EndDate)
	}

	query = query.Group("g.id, g.domain_id, g.f_id")

	if opts.Limit != nil {
		query = query.Limit(*opts.Limit)
	}
	if opts.Offset != nil {
		query = query.Offset(*opts.Offset)
	}

	if err := query.Find(&mGuests).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrGuestsNotFound
		}
		return nil, err
	}

	//todo:ДОДЕЛАТЬ парсинг результатов в структуру и НЕ ЗАБЫТЬ ПРО ORDER - ПО LAST_ACTIVE???
	var guests []domain.Guest

	for _, guest := range *mGuests {
		guests = append(guests, domain.Guest{
			ID:       guest.ID,
			DomainID: guest.DomainID,
			// FirstVisit: ,
		})
	}

}
