package service

import (
	"log/slog"
	"metrika/internal/models"
	"metrika/internal/repository"
	"metrika/internal/tracker"
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
