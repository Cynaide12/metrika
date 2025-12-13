package handlers

import (
	"errors"
	"log/slog"
	"metrika/internal/models"
	"metrika/internal/service"
	response "metrika/lib/api"
	"metrika/lib/logger/sl"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	log     *slog.Logger
	service HandlerService
}

func New(log *slog.Logger, service HandlerService, r chi.Router) *Handler {

	h := &Handler{
		log:     log,
		service: service,
	}

	r.Get("/health", h.Health())
	r.Post("/api/v1/events", h.AddEvent())
	r.Post("/api/v1/sessions", h.NewSession())
	r.Get("/api/v1/metrika/{id}/sessions/online", h.GetCountActiveSessions())

	return h
}

func (h Handler) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		h.log.Debug("HEADERS", slog.Any("HEADERS", r.Header))
		h.log.Debug("REMOTEADDR", slog.Any("REMOTEADDR", r.RemoteAddr))
		h.log.Debug("HOST", slog.Any("HOST", r.Host))

		w.WriteHeader(http.StatusOK)

	}

}

// *EVENTS

func (h Handler) AddEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var fn = "internal.http-server.handlers"

		log := h.log.With("fn", fn)

		var req AddEventsRequest

		if err := render.Decode(r, &req); err != nil {
			log.Error("unable to decode request", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.ErrorWithStatus(response.StatusBadRequest, "bad request"))
			return
		}

		if err := response.ValidateRequest(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.ValidateRequest(validateErr))
			return
		}

		for _, reqEvent := range req.Events {
			event := models.Event{
				SessionID: reqEvent.SessionID,
				Element:   reqEvent.Element,
				// Data:      req.Data,
				Type:      reqEvent.Type,
				Timestamp: reqEvent.Timestamp,
				PageURL:   reqEvent.PageURL,
			}
			//чтобы не блокировать
			go h.service.AddEvent(&event, log)
		}

		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, response.OK())
	}
}

// *SESSIONS

func (h Handler) NewSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var fn = "internal.http-server.handlers.NewSession"

		logger := h.log.With("handlerFn", fn)

		var req CreateNewSessionRequest

		if err := render.Decode(r, &req); err != nil {
			h.log.Error("unable to decode request", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.ErrorWithStatus(response.StatusBadRequest, "bad request"))
			return
		}

		if err := response.ValidateRequest(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.ValidateRequest(validateErr))
			return
		}

		ipAddress := r.Header.Get("X-Forwarded-For")

		session, err := h.service.CreateNewSession(req.FingerprintID, ipAddress, "test.ru")
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			logger.Error("ошибка при создании сессии для юзера", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal server error"))
			return
		}

		render.JSON(w, r, CreateNewSessionResponse{UserId: session.UserID, SessionId: session.ID})
	}
}

//*INFO

func (h Handler) GetCountActiveSessions() http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request){
		var fn = "internal.http-server.handlers.NewSession"

		
		domainId, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil{
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w,r, response.ErrorWithStatus(response.StatusBadRequest, "bad domain id"))
			return
		}
		logger := h.log.With("handlerFn", fn, "domain_id", domainId)

		count, err := h.service.GetCountActiveSessions(uint(domainId))
		if err != nil{
			logger.Error("failed to get count active sessions", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w,r, response.Error("failed to get active sessions"))
			return
		}


		render.JSON(w, r, GetCountActiveSessionsResponse{
			Count: count,
		})

	}
}