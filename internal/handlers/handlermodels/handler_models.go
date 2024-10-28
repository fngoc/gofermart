package handlermodels

// AuthRequest схема запроса для регистрации и авторизации
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// WithdrawRequest схема запроса на списание
type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
