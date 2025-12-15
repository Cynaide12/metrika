package handlers

import (
	"log/slog"
	response "metrika/lib/api"
	"metrika/lib/logger/sl"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// TODO: придумать че делать с роутами тут
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

//*INFO

func (h Handler) GetCountActiveSessions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var fn = "internal.http-server.handlers.NewSession"

		domainId, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.ErrorWithStatus(response.StatusBadRequest, "bad domain id"))
			return
		}
		logger := h.log.With("handlerFn", fn, "domain_id", domainId)

		count, err := h.service.GetCountActiveSessions(uint(domainId))
		if err != nil {
			logger.Error("failed to get count active sessions", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to get active sessions"))
			return
		}

		render.JSON(w, r, GetCountActiveSessionsResponse{
			Count: count,
		})

	}
}
