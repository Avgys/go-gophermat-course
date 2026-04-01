package service

import (
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
)

func NumericToStr(value pgtype.Numeric) string {
	v, err := value.Float64Value()

	if err != nil || !v.Valid {
		v = pgtype.Float8{Valid: true, Float64: 0}
	}

	return strconv.FormatFloat(v.Float64, 'f', 2, 64)
}
