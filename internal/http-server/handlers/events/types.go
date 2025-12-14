package handlers

import (
	"log/slog"
	"metrika/internal/models"
	"time"
)

type EventsService interface {
	AddEvent(e *models.Event, log *slog.Logger)
}

type AddEventsRequest struct {
	Events []AddEventRequest `json:"events" validate:"required"`
}

type AddEventRequest struct {
	SessionID uint                   `json:"session_id" validate:"required"`
	Type      string                 `json:"type" validate:"required"`
	PageURL   string                 `json:"page_url" validate:"required"`
	Element   string                 `json:"element"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}
