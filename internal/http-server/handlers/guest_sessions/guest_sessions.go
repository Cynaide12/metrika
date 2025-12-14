package handlers

import (
	"errors"
	"log/slog"
	"metrika/internal/service"
	response "metrika/lib/api"
	"metrika/lib/logger/sl"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type guestsessions_handler struct {
	log     *slog.Logger
	service GuestSessionsService
}

// *SESSIONS

func (h guestsessions_handler) NewSession() http.HandlerFunc {
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

		render.JSON(w, r, CreateNewSessionResponse{UserId: session.GuestID, SessionId: session.ID})
	}
}