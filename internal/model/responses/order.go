package responses

type Order struct {
	OrderNum     int64  `json:"number"`
	Status       string `json:"status"`
	Accrual      string `json:"accrual"`
	CreatedAtUTC string `json:"uploaded_at"`
}
