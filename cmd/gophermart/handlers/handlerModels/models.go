package handlerModels

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type WithdrawResponse struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
