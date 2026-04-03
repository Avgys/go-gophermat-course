package responses

type WithdrawRs struct {
	OrderNum    int64  `json:"order"`
	Sum         string `json:"sum"`
	ProcessedAt string `json:"processed_at"`
}
