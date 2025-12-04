package service

import (
	"errors"
	"log/slog"
	"metrika/internal/models"
	"metrika/internal/repository"
	"metrika/internal/tracker"
	"metrika/lib/logger/sl"
)

type Service struct {
	log     *slog.Logger
	repo    *repository.Repository
	tracker *tracker.Tracker
}

var (
	ErrNotFound      = repository.ErrNoRows
	ErrAlreadyExists = repository.ErrAlreadyExists
)

func New(repo *repository.Repository, log *slog.Logger, tracker *tracker.Tracker) *Service {
	return &Service{
		log,
		repo,
		tracker,
	}
}

// * EVENTS
func (s *Service) AddEvent(e *models.Event, log *slog.Logger) {
	// var fn = "internal.service.AddEvent"
	// log := s.log.With("session_id", e.SessionID, "user_id", e.UserID, "timestamp", e.Timestamp)
	s.tracker.TrackEvent(*e)
}

func (s *Service) GetDomains(domains *[]models.Domain, opts repository.GetDomainsOptions) error {
	var fn = "internal.service.GetDomains"

	logger := s.log.With("fn", fn)

	if err := s.repo.GetDomains(domains, opts); err != nil {
		logger.Error("failed to get domains from db", sl.Err(err))
		return err
	}

	return nil
}

// * SESSIONS
func (s *Service) CreateNewSession(FingerprintID, IPAddress string, domainUrl string) (models.UserSession, error) {
	var fn = "internal.service.NewSession"

	logger := s.log.With("fn", fn)

	tx := s.repo.GormDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	txStorage := s.repo.WithTx(tx)

	//ищем домен
	var domain models.Domain

	if err := txStorage.GetDomain(&domain, domainUrl); err != nil {
		tx.Rollback()
		if errors.Is(err, repository.ErrNoRows) {
			return models.UserSession{}, ErrNotFound
		}
		logger.Error("ошибка получения домена", sl.Err(err))
		return models.UserSession{}, err
	}

	//ищем юзера по переданному отпечатку
	user, err := txStorage.GetOrCreateUser(FingerprintID, domain.ID)
	if err != nil {
		logger.Error("ошибка получения юзера по f_id", sl.Err(err))
		tx.Rollback()
		return models.UserSession{}, err
	}

	session := models.UserSession{
		UserID:    user.ID,
		IPAddress: IPAddress,
	}

	if err := s.repo.CreateNewSession(&session); err != nil {
		logger.Error("ошибка создания новой сессии", sl.Err(err))
		tx.Rollback()
		return models.UserSession{}, err
	}

	if err := tx.Commit().Error; err != nil {
		logger.Error("ошибка выполнения транзакции", sl.Err(err))

		return models.UserSession{}, err
	}

	return session, nil
}

// func (s *Service) CloseSession(session_id uint) error {
// 	var fn = "internal.service.CloseSession"

// 	logger := s.log.With("fn", fn)

// }
