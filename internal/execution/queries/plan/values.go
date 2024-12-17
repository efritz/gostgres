package plan

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalValuesNode struct {
	fields      []fields.Field
	expressions [][]impls.Expression
}

func NewValues(fields []fields.Field, expressions [][]impls.Expression) LogicalNode {
	return &logicalValuesNode{
		fields:      fields,
		expressions: expressions,
	}
}

func (n *logicalValuesNode) Name() string {
	return "values"
}

func (n *logicalValuesNode) Fields() []fields.Field {
	return n.fields
}

func (n *logicalValuesNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {}
func (n *logicalValuesNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {}
func (n *logicalValuesNode) Optimize(ctx impls.OptimizationContext)                              {}

func (n *logicalValuesNode) EstimateCost() Cost {
	return Cost{} // TODO
}

func (n *logicalValuesNode) Filter() impls.Expression        { return nil }
func (n *logicalValuesNode) Ordering() impls.OrderExpression { return nil }
func (n *logicalValuesNode) SupportsMarkRestore() bool       { return false }

func (n *logicalValuesNode) Build() nodes.Node {
	return nodes.NewValues(n.fields, n.expressions)
}
