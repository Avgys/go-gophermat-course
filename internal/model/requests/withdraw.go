package requests

type WithdrawRq struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}
