package order

import "github.com/samber/lo"

type OrderStatus int

const (
	StatusNew OrderStatus = iota
	StatusProcessing
	StatusInvalid
	StatusProcessed
)

var statusName = map[OrderStatus]string{
	StatusNew:        "NEW",
	StatusProcessing: "PROCESSING",
	StatusInvalid:    "INVALID",
	StatusProcessed:  "PROCESSED",
}

func (ss OrderStatus) String() string {
	name, ok := statusName[ss]

	if !ok {
		name = statusName[StatusInvalid]
	}

	return name
}

func (ss *OrderStatus) Parse(input string) {
	status, ok := lo.FindKey(statusName, input)

	if ok {
		*ss = status
	} else {
		*ss = StatusInvalid
	}
}
