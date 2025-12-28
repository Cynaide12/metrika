package metrika

import (
	"log/slog"
	domain "metrika/internal/domain/analytics"
	"metrika/internal/usecase/metrika"
	response "metrika/pkg/api"
	"metrika/pkg/logger/sl"
	"metrika/pkg/pointers"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type Handler struct {
	log                    *slog.Logger
	getSessions            *metrika.SessionsByRangeDateUseCase
	getCountActiveSessions *metrika.ActiveSessionsUseCase
	getSessionsByInterval  *metrika.SessionsByIntervalUseCase
}

func NewHandler(log *slog.Logger, getSessions *metrika.SessionsByRangeDateUseCase, getCountActiveSessions *metrika.ActiveSessionsUseCase, getSessionsByInterval *metrika.SessionsByIntervalUseCase) *Handler {
	return &Handler{
		log,
		getSessions,
		getCountActiveSessions,
		getSessionsByInterval,
	}
}

type GetGuestSessionByRangeDateResponse struct {
	Sessions *[]domain.GuestSession `json:"sessions"`
	Response response.Response      `json:"response"`
}

func (h *Handler) GetGuestSessionByRangeDate(w http.ResponseWriter, r *http.Request) {

	var opts domain.GuestSessionRepositoryByRangeDateOptions

	domain_id, err := strconv.Atoi(chi.URLParam(r, "domain_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.BadRequest("bad end date"))
		return
	}

	st := r.URL.Query().Get("start_date")
	if st != "" {
		start_date, err := strconv.Atoi(st)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.BadRequest("bad start date"))
			return
		}
		opts.StartDate = pointers.NewTimePointer(time.Unix(int64(start_date), 0))
	}

	ed := r.URL.Query().Get("end_date")

	if ed != "" {
		end_date, err := strconv.Atoi(r.URL.Query().Get("end_date"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.BadRequest("bad end date"))
			return
		}
		opts.EndDate = pointers.NewTimePointer(time.Unix(int64(end_date), 0))
	}

	gi := r.URL.Query().Get("guest_id")

	if gi != "" {
		guest_id, err := strconv.Atoi(gi)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.BadRequest("bad guest id"))
			return
		}
		opts.GuestID = pointers.NewUintPointer(uint(guest_id))
	}

	lt := r.URL.Query().Get("limit")
	if lt != "" {
		limit, err := strconv.Atoi(lt)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.BadRequest("bad limit"))
			return
		}
		opts.Limit = &limit
	}

	of := r.URL.Query().Get("offset")

	if of != "" {
		offset, err := strconv.Atoi(of)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.BadRequest("bad offset"))
			return
		}
		opts.Offset = &offset
	}

	without_active := r.URL.Query().Get("without_active") == "true"
	if without_active {
		opts.WithoutActive = &without_active
	}

	sessions, err := h.getSessions.Execute(r.Context(), uint(domain_id), &opts)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.BadRequest("bad end date"))
		return
	}

	render.JSON(w, r, GetGuestSessionByRangeDateResponse{
		Sessions: sessions,
		Response: response.OK(),
	})
}

type GetCountActiveSessionsResponse struct {
	Online   int64             `json:"online"`
	Response response.Response `json:"response"`
}

func (h *Handler) GetCountActiveSessions(w http.ResponseWriter, r *http.Request) {

	domain_id, err := strconv.Atoi(chi.URLParam(r, "domain_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.BadRequest("bad domain id"))
		return
	}

	count, err := h.getCountActiveSessions.Execute(r.Context(), uint(domain_id))
	if err != nil {
		h.log.Error("ошибка при получении активных сессий домена", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, response.Error("failed to get online guests"))
		return
	}

	render.JSON(w, r, GetCountActiveSessionsResponse{
		Online:   count,
		Response: response.OK(),
	})
}

type GuestSessionsByIntervalResponse struct {
	Response response.Response
	Sessions *[]domain.GuestSessionsByTimeBucket `json:"sessions"`
}

func (h *Handler) GetGuestSessionsByInterval(w http.ResponseWriter, r *http.Request) {

	domain_id, err := strconv.Atoi(chi.URLParam(r, "domain_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.BadRequest("bad domain id"))
		return
	}

	var opts domain.GetVisitsByIntervalOptions

	start_date, err := time.Parse(time.RFC3339, r.URL.Query().Get("start"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.BadRequest("bad start date"))
		return
	}
	opts.Start = start_date

	end_date, err := time.Parse(time.RFC3339, r.URL.Query().Get("end"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.BadRequest("bad end date"))
		return
	}
	opts.End = end_date

	interval, err := strconv.Atoi(r.URL.Query().Get("interval"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.BadRequest("bad interval"))
		return
	}

	opts.IntervalMinutes = interval

	diviser, err := strconv.Atoi(r.URL.Query().Get("diviser"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.BadRequest("bad diviser"))
		return
	}

	opts.IntervalDiviser = diviser

	sessions, err := h.getSessionsByInterval.Execute(r.Context(), uint(domain_id), opts)
	if err != nil {
		h.log.Error("ошибка при получении сессий по интервалам за период", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, response.Error("failed to get o"))
		return
	}

	render.JSON(w, r, GuestSessionsByIntervalResponse{
		Response: response.OK(),
		Sessions: sessions,
	})

}
