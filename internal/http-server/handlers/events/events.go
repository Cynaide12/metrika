package handlers

import (
	"log/slog"
	"metrika/internal/models"
	response "metrika/lib/api"
	"metrika/lib/logger/sl"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type eventshandler struct {
	log     *slog.Logger
	service EventsService
}

func NewAuthHandler(log *slog.Logger, service EventsService, r chi.Router) *eventshandler {

	h := &eventshandler{
		log:     log,
		service: service,
	}

	r.Post("/api/v1/events", h.AddEvent())

	return h
}

// *EVENTS

func (h *eventshandler) AddEvent() http.HandlerFunc {
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
