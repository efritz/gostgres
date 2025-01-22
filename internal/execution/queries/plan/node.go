package plan

import (
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type LogicalNode interface {
	Name() string
	Fields() []fields.Field
	AddFilter(ctx impls.OptimizationContext, filter impls.Expression)
	AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression)
	Optimize(ctx impls.OptimizationContext)
	EstimateCost() impls.NodeCost
	Filter() impls.Expression
	Ordering() impls.OrderExpression

	// TODO: rough implementation
	// https://sourcegraph.com/github.com/postgres/postgres@06286709ee0637ec7376329a5aa026b7682dcfe2/-/blob/src/backend/executor/execAmi.c?L439:59-439:79
	SupportsMarkRestore() bool

	Build() nodes.Node
}

//
//

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
	_ = q.LogicalNode.EstimateCost() // DEBUGGING
	nodes.NewQuery(q.LogicalNode.Build()).Execute(ctx, w)
}
