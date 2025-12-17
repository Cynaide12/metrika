package postgres

import (
	"context"
	domain "metrika/internal/domain/auth"

	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{
		db,
	}
}

func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	db := getDB(ctx, r.db)

	session := UserSession{
		UserID:       s.UserID,
		RefreshToken: s.RefreshToken,
		UserAgent:    s.UserAgent,
	}

	if err := db.Create(&session).Error; err != nil {
		return err
	}

	s.ID = session.ID

	return nil
}

func (r *SessionRepository) ByID(
	ctx context.Context,
	id uint,
) (*domain.Session, error) {
	db := getDB(ctx, r.db)

	var m UserSession
	if err := db.First(&m, id).Error; err != nil {
		return nil, domain.ErrSessionNotFound
	}

	return &domain.Session{
		ID:           m.ID,
		UserID:       m.UserID,
		RefreshToken: m.RefreshToken,
		UserAgent:    m.UserAgent,
	}, nil
}

func (r *SessionRepository) Update(
	ctx context.Context,
	s *domain.Session,
) error {
	db := getDB(ctx, r.db)

	return db.Model(&UserSession{}).
		Where("id = ?", s.ID).
		Update("refresh_token", s.RefreshToken).
		Error
}
