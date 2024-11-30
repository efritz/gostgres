package nodes

import (
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type logicalJoinNode struct {
	left     LogicalNode
	right    LogicalNode
	filter   impls.Expression
	fields   []fields.Field
	strategy logicalJoinStrategy
}

var _ LogicalNode = &logicalJoinNode{}

func NewJoin(left LogicalNode, right LogicalNode, condition impls.Expression) LogicalNode {
	return &logicalJoinNode{
		left:     left,
		right:    right,
		filter:   condition,
		fields:   append(left.Fields(), right.Fields()...),
		strategy: nil,
	}
}

func (n *logicalJoinNode) Name() string {
	return ""
}

func (n *logicalJoinNode) Fields() []fields.Field {
	return slices.Clone(n.fields)
}

func (n *logicalJoinNode) AddFilter(ctx impls.OptimizationContext, filterExpression impls.Expression) {
	n.filter = expressions.UnionFilters(n.filter, filterExpression)
}

func (n *logicalJoinNode) AddOrder(ctx impls.OptimizationContext, orderExpression impls.OrderExpression) {
	lowerOrder(ctx, orderExpression, n.left, n.right)
}

func (n *logicalJoinNode) Optimize(ctx impls.OptimizationContext) {
	// NOTE: Outer fields depend on nested loop join strategy
	// Merge and hash joins won't have have LHS rows available to RHS

	if n.filter != nil {
		n.filter = n.filter.Fold()
		lowerFilter(ctx, n.filter, n.left)
		lowerFilter(ctx.AddOuterFields(n.left.Fields()), n.filter, n.right)
	}

	n.left.Optimize(ctx)
	n.right.Optimize(ctx.AddOuterFields(n.left.Fields()))
	n.filter = expressions.FilterDifference(n.filter, expressions.UnionFilters(n.left.Filter(), n.right.Filter()))
	n.strategy = &logicalNestedLoopJoinStrategy{n: n}
}

func (n *logicalJoinNode) Filter() impls.Expression {
	return expressions.UnionFilters(n.filter, n.left.Filter(), n.right.Filter())
}

func (n *logicalJoinNode) Ordering() impls.OrderExpression {
	if n.strategy == nil {
		panic("No strategy set - optimization required before ordering can be determined")
	}

	return n.strategy.Ordering()
}

func (n *logicalJoinNode) SupportsMarkRestore() bool {
	return false
}

func (n *logicalJoinNode) Build() Node {
	node := &joinNode{
		left:   n.left.Build(),
		right:  n.right.Build(),
		filter: n.filter,
		fields: n.fields,
	}

	node.strategy = n.strategy.Build(node)
	return node
}

//
//

type joinNode struct {
	left     Node
	right    Node
	filter   impls.Expression
	fields   []fields.Field
	strategy joinStrategy
}

var _ Node = &joinNode{}

func (n *joinNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("join using %s", n.strategy.Name())
	n.left.Serialize(w.Indent())
	w.WritefLine("with")
	n.right.Serialize(w.Indent())

	if n.filter != nil {
		w.WritefLine("on %s", n.filter)
	}
}

func (n *joinNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	if n.strategy == nil {
		panic("No strategy set - optimization required before scanning can be performed")
	}

	return n.strategy.Scanner(ctx)
}
