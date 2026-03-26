package model

type AccrualResponse struct {
	Order   int    `json:"order"`
	Accrual int    `json:"accrual"`
	Status  string `json:"status"`
}
