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
}

func NewHandler(log *slog.Logger, getSessions *metrika.SessionsByRangeDateUseCase, getCountActiveSessions *metrika.ActiveSessionsUseCase) *Handler {
	return &Handler{
		log,
		getSessions,
		getCountActiveSessions,
	}
}

type GetGuestSessionByRangeDateResponse struct {
	Sessions *[]domain.GuestSession
	Response response.Response
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
	Online int64 `json:"online"`
	Response response.Response
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
		Online: count,
		Response: response.OK(),
	})
}
