package analytics

import (
	"context"
	"errors"
	"log/slog"
	domain "metrika/internal/domain/analytics"
	"metrika/internal/usecase/analytics"
	response "metrika/pkg/api"
	"metrika/pkg/logger/sl"
	"net/http"
	"strconv"

	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	log          *slog.Logger
	events       *analytics.CollectEventsUseCase
	sessions     *analytics.GetGuestSessionUseCase
	recordEvents *analytics.CollectRecordEventsUseCase
}

type CollectEventsRequest struct {
	Events []CollectEventRequest `json:"events" validate:"required"`
}

type CollectEventRequest struct {
	SessionID uint                   `json:"session_id" validate:"required"`
	Type      string                 `json:"type" validate:"required"`
	PageURL   string                 `json:"page_url" validate:"required"`
	Element   string                 `json:"element"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

func NewHandler(log *slog.Logger,
	events *analytics.CollectEventsUseCase,
	sessions *analytics.GetGuestSessionUseCase,
	recordEvents *analytics.CollectRecordEventsUseCase) *Handler {
	return &Handler{
		log,
		events,
		sessions,
		recordEvents,
	}
}

func (h *Handler) AddEvent(w http.ResponseWriter, r *http.Request) {
	var req CollectEventsRequest
	if err := render.Decode(r, &req); err != nil {
		http.Error(w, "unable to decode request", http.StatusBadRequest)
		return
	}

	if err := response.ValidateRequest(req); err != nil {
		validateErr := err.(validator.ValidationErrors)
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.ValidationError(validateErr))
		return
	}

	var events []domain.Event
	for _, event := range req.Events {
		e := domain.Event{
			SessionID: event.SessionID,
			Type:      event.Type,
			Element:   event.Element,
			PageURL:   event.PageURL,
			Timestamp: event.Timestamp,
			Data:      event.Data,
		}
		events = append(events, e)
	}

	go h.events.Execute(context.Background(), &events)

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response.OK())
}

type AddRecordEventsRequest struct {
	Events *[]domain.RecordEvent `json:"events" validate:"required"`
}


//TODO: вынести в обработку сохранений по пачкам в воркер
func (h *Handler) AddRecordEvents(w http.ResponseWriter, r *http.Request) {

	session_id, err := strconv.Atoi(chi.URLParam(r, "session_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.BadRequest("invalid session id"))
		return
	}

	var req AddRecordEventsRequest
	if err := render.Decode(r, &req); err != nil {
		http.Error(w, "unable to decode request", http.StatusBadRequest)
		return
	}

	if err := response.ValidateRequest(req); err != nil {
		validateErr := err.(validator.ValidationErrors)
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.ValidationError(validateErr))
		return
	}

	if err := h.recordEvents.Execute(r.Context(), req.Events, uint(session_id)); err != nil {
		if errors.Is(err, domain.ErrSessionsNotFound) {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.BadRequest("session not found"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, response.Error("failed to add record events"))
		return
	}

	w.WriteHeader(http.StatusOK)
	render.JSON(w,r, response.OK())
}

type CreateNewSessionRequest struct {
	FingerprintID string `json:"f_id"`
}

type CreateNewSessionResponse struct {
	UserId    uint `json:"m_u_id"`
	SessionId uint `json:"m_s_id"`
}

func (h *Handler) CreateGuestSession(w http.ResponseWriter, r *http.Request) {
	var req CreateNewSessionRequest
	if err := render.Decode(r, &req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := response.ValidateRequest(req); err != nil {
		validateErr := err.(validator.ValidationErrors)
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.ValidationError(validateErr))
		return
	}

	logger := h.log.With("fingerprint_id", req.FingerprintID)

	ipAddress := r.Header.Get("X-Forwarded-For")

	//TODO: не забыть реализовать разные домены
	session, err := h.sessions.Execute(r.Context(), req.FingerprintID, ipAddress, "test.ru")
	if err != nil {
		logger.Error("ошибка создания гостевой сессии", sl.Err(err))
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, CreateNewSessionResponse{
		UserId:    session.GuestID,
		SessionId: session.ID,
	})
}
