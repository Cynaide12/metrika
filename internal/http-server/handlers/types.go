package handlers


type HandlerService interface {
	GetCountActiveSessions(domain_id uint) (int64, error)
}


type GetCountActiveSessionsResponse struct {
	Count int64 `json:"count"`
}