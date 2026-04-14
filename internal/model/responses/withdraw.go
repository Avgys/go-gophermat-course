package responses

type WithdrawRs struct {
	OrderNum    string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type WithdrawDeltaRs struct {
	Modified  bool    `json:"modified"`
	NewAmount float32 `json:"new_amount"`
	OldAmount float32 `json:"old_amount"`
}
