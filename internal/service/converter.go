package service

import (
	"github.com/jackc/pgx/v5/pgtype"
)

func NumericToFloat(value pgtype.Numeric) float64 {
	v, err := value.Float64Value()

	if err != nil || !v.Valid {
		v = pgtype.Float8{Valid: true, Float64: 0}
	}

	return v.Float64
}
