package service

import (
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

func New(repo *repository.Repository, log *slog.Logger, tracker *tracker.Tracker) *Service {
	return &Service{
		log,
		repo,
		tracker,
	}
}

func (s *Service) AddEvent(e *models.Event, log *slog.Logger) {
	// var fn = "internal.service.AddEvent"
	// log := s.log.With("session_id", e.SessionID, "user_id", e.UserID, "timestamp", e.Timestamp)
	s.tracker.TrackEvent(*e)
}


func (s *Service) CreateNewSession(FingerprintID, IPAddress string, domain string, log *slog.Logger) error {
	var fn = "internal.service.NewSession"

	tx := s.repo.WithTx(s.repo.GormDB)

	txStorage := tx.GormDB.Begin()


	//ищем юзера по переданному отпечатку
	user, err := s.repo.GetOrCreateUser(FingerprintID)
	if err != nil {
		s.log.Error("ошибка получения юзера по f_id", sl.Err(err))
		return err
	}



	session := models.UserSession{
		FingerprintID,
		IPAddress,
	}

	if err := s.repo.CreateNewSession(&session); err != nil{

	}

}