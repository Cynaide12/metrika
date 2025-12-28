package postgres

import (
	"context"
	"errors"
	domain "metrika/internal/domain/analytics"
	"time"

	"gorm.io/gorm"
)

type GuestSessionRepository struct {
	db *gorm.DB
}

func NewGuestSessionRepository(db *gorm.DB) *GuestSessionRepository {
	return &GuestSessionRepository{db}
}

func (d *GuestSessionRepository) Create(ctx context.Context, session *domain.GuestSession) error {
	db := getDB(ctx, d.db)

	mSessions := GuestSession{
		IPAddress:  session.IPAddress,
		GuestID:    session.GuestID,
		Active:     session.Active,
		LastActive: session.LastActive,
		EndTime:    session.EndTime,
	}

	if err := db.Model(&GuestSession{}).Create(&mSessions).Error; err != nil {
		return err
	}

	session.ID = mSessions.ID

	return nil
}

func (d *GuestSessionRepository) CreateSessions(ctx context.Context, sessions *[]domain.GuestSession) ([]domain.GuestSession, error) {
	db := getDB(ctx, d.db)

	var mSessions []GuestSession
	for _, session := range *sessions {
		mSessions = append(mSessions, GuestSession{
			IPAddress:  session.IPAddress,
			GuestID:    session.GuestID,
			Active:     session.Active,
			LastActive: session.LastActive,
			EndTime:    session.EndTime,
		})
	}

	if err := db.Model(&GuestSession{}).Create(&mSessions).Error; err != nil {
		return nil, err
	}

	var dDessions []domain.GuestSession
	for _, session := range mSessions {
		dDessions = append(dDessions, domain.GuestSession{
			ID:         session.ID,
			IPAddress:  session.IPAddress,
			GuestID:    session.GuestID,
			Active:     session.Active,
			LastActive: session.LastActive,
			EndTime:    session.EndTime,
		})
	}

	return dDessions, nil
}

func (d *GuestSessionRepository) GetCountActiveSessions(ctx context.Context, domain_id uint) (int64, error) {
	db := getDB(ctx, d.db)

	res := db.Exec("SELECT * FROM guest_sessions s LEFT JOIN guests u ON u.id=s.guest_id WHERE s.active = true AND u.domain_id=?", domain_id)

	if res.Error != nil {
		return 0, res.Error
	}

	return res.RowsAffected, nil
}

func (d *GuestSessionRepository) GetVisitsByInterval(
	ctx context.Context,
	domain_id uint,
	opts domain.GetVisitsByIntervalOptions,
) (*[]domain.GuestSessionsByTimeBucket, error) {

	var buckets []domain.GuestSessionsByTimeBucket

	query := `
        SELECT 
            DATE_TRUNC('hour', created_at) + 
            INTERVAL '? min' * FLOOR(EXTRACT(minute FROM created_at) / ?) as time_bucket,
            COUNT(*) as visits,
            COUNT(DISTINCT guest_id) as uniques
        FROM guest_sessions
        WHERE created_at BETWEEN ? AND ? 
        GROUP BY 1
        ORDER BY time_bucket
    `

	err := d.db.Raw(query,
		opts.IntervalMinutes,
		opts.IntervalDiviser,
		opts.Start,
		opts.End,
	).Scan(&buckets).Error

	return &buckets, err
}

func (d *GuestSessionRepository) ByRangeDate(ctx context.Context, opts domain.GuestSessionRepositoryByRangeDateOptions) (*[]domain.GuestSession, error) {
	db := getDB(ctx, d.db)

	var mSessions []GuestSession

	query := db.Debug().Model(GuestSession{})

	if opts.StartDate != nil {
		query.Where("NOT created_at < ?", opts.StartDate)
	}
	if opts.EndDate != nil {
		query.Where("NOT created_at > ?", opts.EndDate)
	}
	if opts.WithoutActive != nil {
		query.Where("active = false")
	}
	if opts.GuestID != nil {
		query.Where("guest_id = ?", opts.GuestID)
	}
	if opts.Limit != nil {
		query.Limit(*opts.Limit)
	}
	if opts.Offset != nil {
		query.Offset(*opts.Offset)
	}

	if err := query.Order("id ASC").Debug().Find(&mSessions).Error; err != nil {
		if errors.Is(err, domain.ErrSessionsNotFound) {
			return nil, domain.ErrSessionsNotFound
		}
		return nil, err
	}

	var sessions []domain.GuestSession

	for _, session := range mSessions {
		sessions = append(sessions, domain.GuestSession{
			ID:         session.ID,
			GuestID:    session.GuestID,
			EndTime:    session.EndTime,
			LastActive: session.LastActive,
			Active:     session.Active,
		})
	}

	return &sessions, nil
}

// TODO: протестить
func (d *GuestSessionRepository) LastActiveByGuestId(ctx context.Context, guest_id uint) (*domain.GuestSession, error) {
	db := getDB(ctx, d.db)

	var mSession GuestSession

	if err := db.Debug().Model(GuestSession{}).Where("active = true AND guest_id = ?", guest_id).Last(&mSession).Error; err != nil {
		if errors.Is(err, domain.ErrLastActiveSessionNotFound) {
			return nil, domain.ErrLastActiveSessionNotFound
		}

		return nil, err
	}

	session := domain.GuestSession{ID: mSession.ID,
		GuestID:    mSession.GuestID,
		IPAddress:  mSession.IPAddress,
		LastActive: mSession.LastActive,
		EndTime:    mSession.EndTime,
	}

	return &session, nil
}

func (d *GuestSessionRepository) SetLastActive(ctx context.Context, session_ids []uint, last_active time.Time) error {

	db := getDB(ctx, d.db)

	if err := db.Model(&GuestSession{}).Where("id IN ?", session_ids).Update("last_active", last_active).Error; err != nil {
		return err
	}

	return nil
}

func (d *GuestSessionRepository) GetStaleSessions(ctx context.Context, limit int) (*[]domain.GuestSession, error) {
	db := getDB(ctx, d.db)

	var sessions []GuestSession

	if err := db.Raw(`
	SELECT s.id, s.last_active FROM guest_sessions s WHERE s.active = true AND s.last_active < NOW() - INTERVAL  '25 minutes'
	AND NOT EXISTS 
	(SELECT 1 FROM events e WHERE e.session_id=s.id AND e.timestamp > NOW() - INTERVAL '30 minutes') 
	LIMIT $1
	`, limit).Scan(&sessions).Error; err != nil {
		return nil, err
	}

	var dSessions []domain.GuestSession
	for _, session := range sessions {
		dSession := domain.GuestSession{
			ID:         session.ID,
			GuestID:    session.GuestID,
			IPAddress:  session.IPAddress,
			Active:     session.Active,
			LastActive: session.LastActive,
			EndTime:    session.EndTime,
		}

		dSessions = append(dSessions, dSession)
	}

	return &dSessions, nil
}

func (d *GuestSessionRepository) CloseSessions(ctx context.Context, session_ids []uint) error {
	db := getDB(ctx, d.db)

	if err := db.Exec("UPDATE guest_sessions SET active = false, end_time = NOW() WHERE id = ANY($1)", session_ids).Error; err != nil {
		return err
	}
	return nil
}
