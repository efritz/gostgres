package queries

import (
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type Query interface {
	Execute(ctx impls.Context, w protocol.ResponseWriter)
}

type NodeQuery struct {
	Node Node
}

var _ Query = &NodeQuery{}

func NewQuery(n Node) *NodeQuery {
	return &NodeQuery{
		Node: n,
	}
}

func (q *NodeQuery) Execute(ctx impls.Context, w protocol.ResponseWriter) {
	q.Node.Optimize()

	scanner, err := q.Node.Scanner(ctx)
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

func Evaluate(ctx impls.Context, expr impls.Expression, row rows.Row) (any, error) {
	return expr.ValueFrom(ctx, rows.CombineRows(row, ctx.OuterRow()))
}
