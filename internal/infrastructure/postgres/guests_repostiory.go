package postgres

import (
	"context"
	"errors"
	"fmt"
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

func (r *GuestsRepository) Find(ctx context.Context, opts domain.FindGuestsOptions) ([]domain.Guest, int64, error) {
	var allowedOrders = map[string]string{
		"first_visit":           "first_visit",
		"last_visit":            "last_visit",
		"guest_id":              "g.id",
		"total_seconds_on_site": "total_seconds_on_site",
		"is_online":             "is_online",
		"sessions_count":        "sessions_count",
	}

	db := getDB(ctx, r.db)

	var mGuests *[]GuestDTO

	query := db.Table("guests g").Debug().
		Select(`
	g.id,
	g.domain_id, 
	g.f_id,
	MIN(gs.created_at) AS first_visit,
	MAX(gs.created_at) AS last_visit,
	COUNT(gs.id) AS sessions_count,
	SUM(CASE 
			WHEN gs.end_time IS NOT NULL
			THEN EXTRACT(EPOCH FROM (gs.end_time - gs.created_at))
			ELSE EXTRACT(EPOCH FROM (gs.last_active - gs.created_at))
		END
	) AS total_seconds_on_site,
	EXISTS(
	SELECT 1 FROM guest_sessions ss
	WHERE ss.guest_id=g.id AND ss.active=true AND ss.end_time IS NULL
	) as is_online
	`).Joins(`
	LEFT JOIN guest_sessions gs ON g.id=gs.guest_id
	`).Where("g.domain_id=?", opts.DomainID)

	if opts.StartDate != nil && opts.EndDate != nil {
		query = query.Where("g.id IN (SELECT gs2.guest_id FROM guest_sessions gs2 WHERE gs2.created_at >= ? AND gs2.created_at <= ?)", opts.StartDate, opts.EndDate)
	} else if opts.StartDate != nil {
		query = query.Where("g.id IN (SELECT gs2.guest_id FROM guest_sessions gs2 WHERE gs2.created_at >= ?)", opts.StartDate)
	} else if opts.EndDate != nil {
		query = query.Where("g.id IN (SELECT gs2.guest_id FROM guest_sessions gs2 WHERE gs2.created_at <= ?)", opts.EndDate)
	}

	query = query.Group("g.id, g.domain_id, g.f_id")

	countQuery := *query

	var count int64
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if opts.Limit != nil {
		query = query.Limit(*opts.Limit)
	}
	if opts.Offset != nil {
		query = query.Offset(*opts.Offset)
	}

	if opts.Order != nil && opts.OrderType != nil {
		columnName, ok := allowedOrders[*opts.Order]
		if !ok {
			return nil, 0, domain.ErrFindGuestsOrderNotAllowed
		}
		query.Order(fmt.Sprintf("%s %s", columnName, *opts.OrderType))
	} else {
		query.Order("last_visit ASC")
	}

	if err := query.Find(&mGuests).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, domain.ErrGuestsNotFound
		}
		return nil, 0, err
	}

	var guests []domain.Guest

	for _, guest := range *mGuests {
		guests = append(guests, guest.ToDomain())
	}

	return guests, count, nil
}

func (r *GuestsRepository) ByID(ctx context.Context, guest_id uint) (*domain.Guest, error) {
	db := getDB(ctx, r.db)

	var mGuest GuestDTO

	if err := db.Table("guests g").Select(`
	g.id, 
	g.domain_id, 
	g.f_id, 
	SUM(
		CASE
			WHEN gs.end_time IS NOT NULL
			THEN EXTRACT(EPOCH FROM (gs.end_time - gs.created_at))
			ELSE EXTRACT (EPOCH FROM (gs.last_active - gs.created_at))
		END	
		) AS total_seconds_on_site,
	MIN(gs.created_at) AS first_visit,
	MAX(gs.created_at) AS last_visit,
	COUNT(gs.id) AS sessions_count,
	EXISTS(SELECT 1 FROM guest_sessions ss WHERE ss.guest_id=g.id AND active=true AND end_time IS NULL) AS is_online
		`).Joins("LEFT JOIN guest_sessions gs ON g.id=gs.guest_id").Where("g.id=?", guest_id).Group("g.id").First(&mGuest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrGuestsNotFound
		}
		return nil, err
	}

	guest := mGuest.ToDomain()

	return &guest, nil
}
