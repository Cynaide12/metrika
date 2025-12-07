package handlers

import (
	"log/slog"
	"metrika/internal/models"
	"time"
)

type HandlerService interface {
	AddEvent(e *models.Event, log *slog.Logger)
	CreateNewSession(FingerprintID, IPAddress string, domainUrl string) (models.UserSession, error)
}

type AddEventRequest struct {
	SessionID uint                 `json:"session_id" validate:"required"`
	Type      string                 `json:"type" validate:"required"`
	PageURL   string                 `json:"page_url" validate:"required"`
	Element   string                 `json:"element"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type CreateNewSessionRequest struct {
	FingerprintID string `json:"f_id"`
}

type CreateNewSessionResponse struct{
	UserId uint `json:"m_u_id"`
	SessionId uint `json:"m_s_id"`
}