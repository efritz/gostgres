package opt

import (
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type NodeQuery struct {
	LogicalNode
}

func NewQuery(n LogicalNode) queries.Query {
	return &NodeQuery{
		LogicalNode: n,
	}
}

func (q *NodeQuery) Execute(ctx impls.ExecutionContext, w protocol.ResponseWriter) {
	q.LogicalNode.Optimize(ctx.OptimizationContext())
	nodes.NewQuery(q.LogicalNode.Build()).Execute(ctx, w)
}
