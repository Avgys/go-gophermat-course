package order

import (
	"time"
)

type Order struct {
	Id           int64
	OrderNum     int64
	Status       OrderStatus
	Accrual      int
	CreatedAtUTC time.Time
}
