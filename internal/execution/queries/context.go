package queries

import (
	"github.com/efritz/gostgres/internal/execution"
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

func Evaluate(ctx execution.Context, expr expressions.Expression, row shared.Row) (any, error) {
	return expr.ValueFrom(ctx, shared.CombineRows(row, ctx.OuterRow()))
}
