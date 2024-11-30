package nodes

import (
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type NodeQuery struct {
	Node
}

func NewQuery(n Node) queries.Query {
	return &NodeQuery{
		Node: n,
	}
}

func (q *NodeQuery) Execute(ctx impls.ExecutionContext, w protocol.ResponseWriter) {
	scanner, err := q.Scanner(ctx)
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
