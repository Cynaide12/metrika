package handlers

import "metrika/internal/models"

type GuestSessionsService interface {
	CreateNewSession(FingerprintID, IPAddress string, domainUrl string) (models.GuestSession, error)
}

type CreateNewSessionRequest struct {
	FingerprintID string `json:"f_id"`
}

type CreateNewSessionResponse struct {
	UserId    uint `json:"m_u_id"`
	SessionId uint `json:"m_s_id"`
}