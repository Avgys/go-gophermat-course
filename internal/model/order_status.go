package model

import "github.com/samber/lo"

type OrderStatus int

const (
	New OrderStatus = iota
	Processing
	Invalid
	Processed
)

var statusName = map[OrderStatus]string{
	New:        "NEW",
	Processing: "PROCESSING",
	Invalid:    "INVALID",
	Processed:  "PROCESSED",
}

func (ss OrderStatus) String() string {
	return statusName[ss]
}

func (ss *OrderStatus) Parse(input string) {
	status, ok := lo.FindKey(statusName, input)

	if ok {
		*ss = status
	} else {
		*ss = Invalid
	}
}
