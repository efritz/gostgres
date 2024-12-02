package plan

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/types"
)

type logicalExplain struct {
	LogicalNode
}

func NewExplain(n LogicalNode) *logicalExplain {
	return &logicalExplain{
		LogicalNode: n,
	}
}

func (n *logicalExplain) Name() string {
	return "EXPLAIN"
}

var queryPlanField = fields.NewField("", "query plan", types.TypeText, fields.NonInternalField)

func (n *logicalExplain) Fields() []fields.Field {
	return []fields.Field{queryPlanField}
}

func (n *logicalExplain) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {} // top-level
func (n *logicalExplain) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {} // top-level
func (n *logicalExplain) Filter() impls.Expression                                            { return nil }
func (n *logicalExplain) Ordering() impls.OrderExpression                                     { return nil }
func (n *logicalExplain) SupportsMarkRestore() bool                                           { return false }

func (n *logicalExplain) Build() nodes.Node {
	return nodes.NewExplain(n.LogicalNode.Build(), n.Fields())
}
