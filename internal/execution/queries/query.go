package queries

import (
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
)

type Query interface {
	Execute(ctx impls.ExecutionContext, w protocol.ResponseWriter)
}

func Evaluate(ctx impls.ExecutionContext, expr impls.Expression, row rows.Row) (any, error) {
	return expr.ValueFrom(ctx, rows.CombineRows(row, ctx.OuterRow()))
}

func EvaluateExpressions(ctx impls.ExecutionContext, expressions []impls.Expression, row rows.Row) (values []any, _ error) {
	for _, expression := range expressions {
		value, err := Evaluate(ctx, expression, row)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}
