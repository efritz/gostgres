package queries

import (
	"github.com/efritz/gostgres/internal/eval"
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type Context = eval.Context

func Evaluate(ctx Context, expr expressions.Expression, row shared.Row) (any, error) {
	return expr.ValueFrom(ctx, shared.CombineRows(row, ctx.OuterRow()))
}
