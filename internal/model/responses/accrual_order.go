package responses

type AccrualOrder struct {
	OrderNum string  `json:"order"`
	Accrual  float32 `json:"accrual"`
	Status   string  `json:"status"`
}
