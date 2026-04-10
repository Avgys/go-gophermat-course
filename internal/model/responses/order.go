package responses

type Order struct {
	OrderNum     string  `json:"number"`
	Status       string  `json:"status"`
	Accrual      float64 `json:"accrual"`
	CreatedAtUTC string  `json:"uploaded_at"`
}
