package queries

import (
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type Query interface {
	Execute(ctx impls.ExecutionContext, w protocol.ResponseWriter)
}

type NodeQuery struct {
	LogicalNode LogicalNode
}

var _ Query = &NodeQuery{}

func NewQuery(n LogicalNode) *NodeQuery {
	return &NodeQuery{
		LogicalNode: n,
	}
}

func (q *NodeQuery) Execute(ctx impls.ExecutionContext, w protocol.ResponseWriter) {
	q.LogicalNode.Optimize(ctx.OptimizationContext())

	scanner, err := q.LogicalNode.Build().Scanner(ctx)
	if err != nil {
		w.Error(err)
		return
	}

	if err := scan.VisitRows(scanner, func(row rows.Row) (bool, error) {
		w.SendRow(row)
		return true, nil
	}); err != nil {
		w.Error(err)
		return
	}

	w.Done()
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
