package queries

import (
	"github.com/efritz/gostgres/internal/protocol"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
)

type Query interface {
	Execute(ctx Context, w protocol.ResponseWriter)
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

func (q *NodeQuery) Execute(ctx Context, w protocol.ResponseWriter) {
	q.Node.Optimize()

	scanner, err := q.Node.Scanner(ctx)
	if err != nil {
		w.Error(err)
		return
	}

	if err := scan.VisitRows(scanner, func(row shared.Row) (bool, error) {
		w.SendRow(row)
		return true, nil
	}); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}