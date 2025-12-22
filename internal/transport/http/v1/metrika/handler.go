package metrika

import (
	"log/slog"
	domain "metrika/internal/domain/analytics"
	"metrika/internal/usecase/metrika"
	"metrika/pkg/pointers"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	log      *slog.Logger
	sessions *metrika.SessionsByRangeDateUseCase
}

func NewMetrikaHandler(log *slog.Logger, sessions *metrika.SessionsByRangeDateUseCase) *Handler {
	return &Handler{
		log,
		sessions,
	}
}

// TODO: ДОДЕЛАТЬ
func (h *Handler) GetGuestSessionByRangeDate(r *http.Request, w http.ResponseWriter) {

	start_date, err := strconv.Atoi(r.URL.Query().Get("start_date"))
	if err != nil {
		http.Error(w, "bad start date", http.StatusBadRequest)
		return
	}

	end_date, err := strconv.Atoi(r.URL.Query().Get("end_date"))
	if err != nil {
		http.Error(w, "bad end date", http.StatusBadRequest)
		return
	}

	domain_id, err := strconv.Atoi(r.URL.Query().Get("domain_id"))
	if err != nil {
		http.Error(w, "bad domain id", http.StatusBadRequest)
		return
	}

	opts := domain.GuestSessionRepositoryByRangeDateOptions{
		StartDate: pointers.NewTimePointer(time.Unix(int64(start_date), 0)),
		EndDate:   pointers.NewTimePointer(time.Unix(int64(end_date), 0)),
	}

	h.sessions.Execute(r.Context(), uint(domain_id), &opts)

}
