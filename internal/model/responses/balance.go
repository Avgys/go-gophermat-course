package responses

type Balance struct {
	CurrentSum float64 `json:"current"`
	Withdrawn  float64 `json:"withdrawn"`
}
