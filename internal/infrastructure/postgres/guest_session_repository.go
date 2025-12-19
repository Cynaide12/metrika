package postgres

import (
	"context"
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

// TODO: доделать
func (d *GuestSessionRepository) GetCountActiveSessions(ctx context.Context, domain_id uint) (int64, error) {
	db := getDB(ctx, d.db)

	var count int64

	if err := db.Debug().Model(&domain.GuestSession{}).Exec("SELECT * FROM guest_sessions s LEFT JOIN guests u ON u.id=s.guest_id WHERE s.active = true AND u.domain_id=?", domain_id).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (d *GuestSessionRepository) SetLastActive(ctx context.Context, session_ids map[uint]struct{}, last_active time.Time) error {

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

	if err := db.Exec("UPDATE user_sessions SET active = false, end_time = NOW() WHERE id = ANY($1)", session_ids).Error; err != nil {
		return err
	}
	return nil
}
