package response

type Balance struct {
	CurrentSum string `json:"current"`
	Withdrawn  string `json:"withdrawn"`
}
