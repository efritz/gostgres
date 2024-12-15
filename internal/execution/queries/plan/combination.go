package plan

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/execution/queries/plan/util"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalCombinationNode struct {
	left             LogicalNode
	right            LogicalNode
	fields           []fields.Field
	groupedRowFilter nodes.GroupedRowFilterFunc
	distinct         bool
}

func NewIntersect(left LogicalNode, right LogicalNode, distinct bool) (LogicalNode, error) {
	return newCombination(left, right, intersectFilter, distinct)
}

func NewExcept(left LogicalNode, right LogicalNode, distinct bool) (LogicalNode, error) {
	return newCombination(left, right, exceptFilter, distinct)
}

func newCombination(left LogicalNode, right LogicalNode, groupedRowFilter nodes.GroupedRowFilterFunc, distinct bool) (LogicalNode, error) {
	leftFields := left.Fields()
	rightFields := right.Fields()

	if len(leftFields) != len(rightFields) {
		return nil, fmt.Errorf("unexpected combination columns")
	}
	for i, leftField := range leftFields {
		if leftField.Type() != rightFields[i].Type() {
			// TODO - refine type if possible
			return nil, fmt.Errorf("unexpected combination types")
		}
	}

	return &logicalCombinationNode{
		left:             left,
		right:            right,
		fields:           leftFields,
		groupedRowFilter: groupedRowFilter,
		distinct:         distinct,
	}, nil
}

func (n *logicalCombinationNode) Name() string {
	return ""
}

func (n *logicalCombinationNode) Fields() []fields.Field {
	return slices.Clone(n.fields)
}

func (n *logicalCombinationNode) AddFilter(ctx impls.OptimizationContext, filterExpression impls.Expression) {
	util.LowerFilter(ctx, filterExpression, n.left, n.right)
}

func (n *logicalCombinationNode) AddOrder(ctx impls.OptimizationContext, orderExpression impls.OrderExpression) {
	util.LowerOrder(ctx, orderExpression, n.left, n.right)
}

func (n *logicalCombinationNode) Optimize(ctx impls.OptimizationContext) {
	n.left.Optimize(ctx)
	n.right.Optimize(ctx)
}

func (n *logicalCombinationNode) Filter() impls.Expression {
	return n.left.Filter()
}

func (n *logicalCombinationNode) Ordering() impls.OrderExpression { return nil }
func (n *logicalCombinationNode) SupportsMarkRestore() bool       { return false }

func (n *logicalCombinationNode) Build() nodes.Node {
	return nodes.NewCombination(
		n.left.Build(),
		n.right.Build(),
		n.fields,
		n.groupedRowFilter,
		n.distinct,
	)
}

//
//

func intersectFilter(rows []nodes.SourcedRow) bool {
	indexes := map[int]bool{}
	for _, r := range rows {
		indexes[r.Index] = true
	}

	if !indexes[0] || !indexes[1] {
		return false
	}

	return true
}

func exceptFilter(rows []nodes.SourcedRow) bool {
	for _, r := range rows {
		if r.Index == 1 {
			return false
		}
	}

	return true
}
