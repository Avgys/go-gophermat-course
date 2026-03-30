package order

import "github.com/samber/lo"

type OrderStatus int

const (
	StatusNew OrderStatus = iota
	StatusProcessing
	StatusInvalid
	StatusProcessed
)

var StatusName = map[OrderStatus]string{
	StatusNew:        "NEW",
	StatusProcessing: "PROCESSING",
	StatusInvalid:    "INVALID",
	StatusProcessed:  "PROCESSED",
}

func (ss OrderStatus) String() string {
	name, ok := StatusName[ss]

	if !ok {
		name = StatusName[StatusInvalid]
	}

	return name
}

func (ss *OrderStatus) Parse(input string) {
	status, ok := lo.FindKey(StatusName, input)

	if ok {
		*ss = status
	} else {
		*ss = StatusInvalid
	}
}
