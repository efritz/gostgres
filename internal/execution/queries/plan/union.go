package plan

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalUnionNode struct {
	left     LogicalNode
	right    LogicalNode
	fields   []fields.Field
	distinct bool
}

func NewUnion(left LogicalNode, right LogicalNode, distinct bool) (LogicalNode, error) {
	leftFields := left.Fields()
	rightFields := right.Fields()

	if len(leftFields) != len(rightFields) {
		return nil, fmt.Errorf("unexpected union columns")
	}
	for i, leftField := range leftFields {
		if leftField.Type() != rightFields[i].Type() {
			// TODO - refine type if possible
			return nil, fmt.Errorf("unexpected union types")
		}
	}

	return &logicalUnionNode{
		left:     left,
		right:    right,
		fields:   leftFields,
		distinct: distinct,
	}, nil
}

func (n *logicalUnionNode) Name() string {
	return ""
}

func (n *logicalUnionNode) Fields() []fields.Field {
	return slices.Clone(n.fields)
}

func (n *logicalUnionNode) AddFilter(ctx impls.OptimizationContext, filterExpression impls.Expression) {
	lowerFilter(ctx, filterExpression, n.left, n.right)
}

func (n *logicalUnionNode) AddOrder(ctx impls.OptimizationContext, orderExpression impls.OrderExpression) {
	lowerOrder(ctx, orderExpression, n.left, n.right)
}

func (n *logicalUnionNode) Optimize(ctx impls.OptimizationContext) {
	n.left.Optimize(ctx)
	n.right.Optimize(ctx)
}

func (n *logicalUnionNode) Filter() impls.Expression {
	return expressions.FilterIntersection(n.left.Filter(), n.right.Filter())
}

func (n *logicalUnionNode) Ordering() impls.OrderExpression { return nil }
func (n *logicalUnionNode) SupportsMarkRestore() bool       { return false }

func (n *logicalUnionNode) Build() nodes.Node {
	return nodes.NewUnion(n.left.Build(), n.right.Build(), n.fields, n.distinct)
}
