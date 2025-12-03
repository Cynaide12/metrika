package handlers

import (
	"log/slog"
	"metrika/internal/models"
	"time"
)

type HandlerService interface {
	AddEvent(e *models.Event, log *slog.Logger)
}

type AddEventRequest struct {
	SessionID string                 `json:"session_id" validate:"required"`
	Type      string                 `json:"type" validate:"required"`
	UserID    string                 `json:"user_id" validate:"required"`
	PageURL   string                 `json:"page_url" validate:"required"`
	Element   string                 `json:"element"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type CreateNewSessionRequest struct {
	FingerprintID string `json:"f_id"`
}


