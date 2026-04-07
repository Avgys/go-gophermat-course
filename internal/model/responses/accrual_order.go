package responses

type AccrualOrder struct {
	OrderNum string  `json:"order"`
	Accrual  float64 `json:"accrual"`
	Status   string  `json:"status"`
}
