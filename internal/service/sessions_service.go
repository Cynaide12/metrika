package service

import (
	"log/slog"
	"metrika/internal/repository"
	"time"
)

type SessionService struct {
	log            *slog.Logger
	repo           *repository.Repository
	interval time.Duration
}

func NewSessionService(log *slog.Logger, repo *repository.Repository, interval time.Duration) *SessionService {
	return &SessionService{
		log,
		repo,
		interval,
	}
}

func (s SessionService) StartSessionManager(){
	ticker := time.NewTicker(s.interval)

	for range ticker.C{
		
	}
}