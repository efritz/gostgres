package queries

import (
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

func Evaluate(ctx types.Context, expr types.Expression, row shared.Row) (any, error) {
	return expr.ValueFrom(ctx, shared.CombineRows(row, ctx.OuterRow()))
}
