package handlers

import (
	"log/slog"
	"metrika/internal/models"
	response "metrika/lib/api"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	log     *slog.Logger
	service HandlerService
}

func New(log *slog.Logger, service HandlerService) *Handler {
	return &Handler{
		log:     log,
		service: service,
	}
}

func (h Handler) AddEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var fn = "internal.http-server.handlers"

		log := h.log.With("fn", fn)

		var req AddEventRequest

		if err := render.Decode(r, &req); err != nil {
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

		event := models.Event{
			SessionID: req.SessionID,
			UserID:    req.UserID,
			Element:   req.Element,
			Data:      req.Data,
			Type:      req.Type,
			Timestamp: req.Timestamp,
			PageURL:   req.PageURL,
		}

		// if err := h.service.AddEvent(&event, log); err != nil {
		// 	log.Error("failed to add event", sl.Err(err))

		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	render.JSON(w, r, response.ErrorWithStatus(response.StatusError, "failed to add event"))
		// 	return
		// }


		//чтобы не блокировать
		go h.service.AddEvent(&event, log)

		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, response.OK())
	}
}
