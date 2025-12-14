package service

import (
	"errors"
	"log/slog"
	"metrika/internal/models"
	"metrika/internal/repository"
	"metrika/internal/tracker"
	"metrika/lib/logger/sl"
	"time"
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
func (s *Service) CreateNewSession(FingerprintID, IPAddress string, domainUrl string) (models.GuestSession, error) {
	var fn = "internal.service.NewSession"

	logger := s.log.With("fn", fn)

	//ищем домен
	var domain models.Domain

	if err := s.repo.GetDomain(&domain, domainUrl); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return models.GuestSession{}, ErrNotFound
		}
		logger.Error("ошибка получения домена", sl.Err(err))
		return models.GuestSession{}, err
	}

	//ищем юзера по переданному отпечатку
	guest, err := s.repo.GetOrCreateGuest(FingerprintID, domain.ID)
	if err != nil {
		logger.Error("ошибка получения юзера по f_id", sl.Err(err))
		return models.GuestSession{}, err
	}

	session := models.GuestSession{
		GuestID:     guest.ID,
		IPAddress:  IPAddress,
		EndTime:    nil,
		Active:     true,
		LastActive: time.Now(),
	}

	if err := s.repo.CreateNewSession(&session); err != nil {
		logger.Error("ошибка создания новой сессии", sl.Err(err))
		return models.GuestSession{}, err
	}

	return session, nil
}

// *INFO
func (s *Service) GetCountActiveSessions(domain_id uint) (int64, error) {
	var fn = "internal.service.GetCountActiveSessions"

	logger := s.log.With("fn", fn)

	count, err := s.repo.GetCountActiveSessions(domain_id)
	if err != nil {
		logger.Error("failed to get active sessions from db", sl.Err(err))
		return 0, err
	}

	return count, nil
}

// func (s *Service) CloseSession(session_id uint) error {
// 	var fn = "internal.service.CloseSession"

// 	logger := s.log.With("fn", fn)

// }
