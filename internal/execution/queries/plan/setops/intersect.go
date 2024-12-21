package setops

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/execution/queries/nodes/setops"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/execution/queries/plan/cost"
	"github.com/efritz/gostgres/internal/execution/queries/plan/util"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalIntersectNode struct {
	left     plan.LogicalNode
	right    plan.LogicalNode
	fields   []fields.Field
	distinct bool
}

func NewIntersect(left plan.LogicalNode, right plan.LogicalNode, distinct bool) (plan.LogicalNode, error) {
	leftFields := left.Fields()
	rightFields := right.Fields()

	if len(leftFields) != len(rightFields) {
		return nil, fmt.Errorf("unexpected intersect columns")
	}
	for i, leftField := range leftFields {
		if leftField.Type() != rightFields[i].Type() {
			// TODO - refine type if possible
			return nil, fmt.Errorf("unexpected intersect types")
		}
	}

	return &logicalIntersectNode{
		left:     left,
		right:    right,
		fields:   leftFields,
		distinct: distinct,
	}, nil
}

func (n *logicalIntersectNode) Name() string {
	return ""
}

func (n *logicalIntersectNode) Fields() []fields.Field {
	return slices.Clone(n.fields)
}

func (n *logicalIntersectNode) AddFilter(ctx impls.OptimizationContext, filterExpression impls.Expression) {
	util.LowerFilter(ctx, filterExpression, n.left, n.right)
}

func (n *logicalIntersectNode) AddOrder(ctx impls.OptimizationContext, orderExpression impls.OrderExpression) {
	util.LowerOrder(ctx, orderExpression, n.left, n.right)
}

func (n *logicalIntersectNode) Optimize(ctx impls.OptimizationContext) {
	n.left.Optimize(ctx)
	n.right.Optimize(ctx)

	// If left is smaller than right, swap order so we materialize the smaller relation
	if n.left.EstimateCost().EstimatedRows < n.right.EstimateCost().EstimatedRows {
		n.left, n.right = n.right, n.left
	}
}

func (n *logicalIntersectNode) EstimateCost() impls.NodeCost {
	return cost.EstimateIntersectCost(n.left.EstimateCost(), n.right.EstimateCost(), n.distinct)
}

func (n *logicalIntersectNode) Filter() impls.Expression {
	return n.left.Filter()
}

func (n *logicalIntersectNode) Ordering() impls.OrderExpression { return nil }
func (n *logicalIntersectNode) SupportsMarkRestore() bool       { return false }

func (n *logicalIntersectNode) Build() nodes.Node {
	return setops.NewIntersect(n.left.Build(), n.right.Build(), n.fields, n.distinct)
}
