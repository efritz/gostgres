package combination

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/execution/queries/nodes/combination"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/execution/queries/plan/util"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalExceptNode struct {
	left     plan.LogicalNode
	right    plan.LogicalNode
	fields   []fields.Field
	distinct bool
}

func NewExcept(left plan.LogicalNode, right plan.LogicalNode, distinct bool) (plan.LogicalNode, error) {
	leftFields := left.Fields()
	rightFields := right.Fields()

	if len(leftFields) != len(rightFields) {
		return nil, fmt.Errorf("unexpected except columns")
	}
	for i, leftField := range leftFields {
		if leftField.Type() != rightFields[i].Type() {
			// TODO - refine type if possible
			return nil, fmt.Errorf("unexpected except types")
		}
	}

	return &logicalExceptNode{
		left:     left,
		right:    right,
		fields:   leftFields,
		distinct: distinct,
	}, nil
}

func (n *logicalExceptNode) Name() string {
	return ""
}

func (n *logicalExceptNode) Fields() []fields.Field {
	return slices.Clone(n.fields)
}

func (n *logicalExceptNode) AddFilter(ctx impls.OptimizationContext, filterExpression impls.Expression) {
	util.LowerFilter(ctx, filterExpression, n.left, n.right)
}

func (n *logicalExceptNode) AddOrder(ctx impls.OptimizationContext, orderExpression impls.OrderExpression) {
	util.LowerOrder(ctx, orderExpression, n.left, n.right)
}

func (n *logicalExceptNode) Optimize(ctx impls.OptimizationContext) {
	n.left.Optimize(ctx)
	n.right.Optimize(ctx)
}

func (n *logicalExceptNode) Filter() impls.Expression {
	return n.left.Filter()
}

func (n *logicalExceptNode) Ordering() impls.OrderExpression { return nil }
func (n *logicalExceptNode) SupportsMarkRestore() bool       { return false }

func (n *logicalExceptNode) Build() nodes.Node {
	return combination.NewExcept(n.left.Build(), n.right.Build(), n.fields, n.distinct)
}
