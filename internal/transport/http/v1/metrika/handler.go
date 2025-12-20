package metrika

import (
	"log/slog"
	domain "metrika/internal/domain/analytics"
)

type Handler struct {
	log      *slog.Logger
	sessions *domain.GuestSessionRepository
}

func NewMetrikaHandler(log *slog.Logger, sessions *domain.GuestSessionRepository) *Handler {
	return &Handler{
		log,
		sessions,
	}
}

//TODO: ДОДЕЛАТЬ
// func GetGuestSessionByRangeDate(r *http.Request, w http.ResponseWriter) {
	
// 	end_date := 

// }
