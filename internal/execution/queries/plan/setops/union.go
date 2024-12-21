package setops

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/execution/queries/nodes/setops"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/execution/queries/plan/cost"
	"github.com/efritz/gostgres/internal/execution/queries/plan/util"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalUnionNode struct {
	left     plan.LogicalNode
	right    plan.LogicalNode
	fields   []fields.Field
	distinct bool
}

func NewUnion(left plan.LogicalNode, right plan.LogicalNode, distinct bool) (plan.LogicalNode, error) {
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
	util.LowerFilter(ctx, filterExpression, n.left, n.right)
}

func (n *logicalUnionNode) AddOrder(ctx impls.OptimizationContext, orderExpression impls.OrderExpression) {
	util.LowerOrder(ctx, orderExpression, n.left, n.right)
}

func (n *logicalUnionNode) Optimize(ctx impls.OptimizationContext) {
	n.left.Optimize(ctx)
	n.right.Optimize(ctx)
}

func (n *logicalUnionNode) EstimateCost() impls.NodeCost {
	return cost.EstimateUnionCost(n.left.EstimateCost(), n.right.EstimateCost(), n.distinct)
}

func (n *logicalUnionNode) Filter() impls.Expression {
	return expressions.FilterIntersection(n.left.Filter(), n.right.Filter())
}

func (n *logicalUnionNode) Ordering() impls.OrderExpression { return nil }
func (n *logicalUnionNode) SupportsMarkRestore() bool       { return false }

func (n *logicalUnionNode) Build() nodes.Node {
	if !n.distinct {
		return setops.NewAppend(n.left.Build(), n.right.Build(), n.fields)
	}

	return setops.NewUnion(n.left.Build(), n.right.Build(), n.fields)
}
