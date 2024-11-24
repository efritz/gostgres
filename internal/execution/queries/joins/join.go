package joins

import (
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/filter"
	"github.com/efritz/gostgres/internal/execution/queries/order"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type joinNode struct {
	left     queries.Node
	right    queries.Node
	filter   impls.Expression
	fields   []fields.Field
	strategy joinStrategy
}

var _ queries.Node = &joinNode{}

func NewJoin(left queries.Node, right queries.Node, condition impls.Expression) queries.Node {
	return &joinNode{
		left:     left,
		right:    right,
		filter:   condition,
		fields:   append(left.Fields(), right.Fields()...),
		strategy: nil,
	}
}

func (n *joinNode) Name() string {
	return ""
}

func (n *joinNode) Fields() []fields.Field {
	return slices.Clone(n.fields)
}

func (n *joinNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("join using %s", n.strategy.Name())
	n.left.Serialize(w.Indent())
	w.WritefLine("with")
	n.right.Serialize(w.Indent())

	if n.filter != nil {
		w.WritefLine("on %s", n.filter)
	}
}

func (n *joinNode) AddFilter(ctx impls.OptimizationContext, filterExpression impls.Expression) {
	n.filter = expressions.UnionFilters(n.filter, filterExpression)
}

func (n *joinNode) AddOrder(ctx impls.OptimizationContext, orderExpression impls.OrderExpression) {
	order.LowerOrder(ctx, orderExpression, n.left, n.right)
}

func (n *joinNode) Optimize(ctx impls.OptimizationContext) {
	// TODO:
	// - only call optimize once per node
	// - n.filter should be available for strategy, but also calculated after strategy selection

	// NOTE: Outer fields depend on nested loop join strategy
	// Merge and hash joins won't have have LHS rows available to RHS

	if n.filter != nil {
		n.filter = n.filter.Fold()
		// fmt.Printf("Want to lower %s to %T\n", n.filter, n.left)
		filter.LowerFilter(ctx, n.filter, n.left)
		// fmt.Printf("Done.\n\n")
		// fmt.Printf("Want to lower %s to %T\n", n.filter, n.right)
		filter.LowerFilter(ctx.AddOuterFields(n.left.Fields()), n.filter, n.right)
		// fmt.Printf("Done.\n\n")
	}

	n.left.Optimize(ctx)
	n.right.Optimize(ctx.AddOuterFields(n.left.Fields()))
	// fmt.Printf("Left filter (%T: %s): %s\nRight filter (%T: %s): %s\n\n", n.left, n.left, n.left.Filter(), n.right, n.right, n.right.Filter())
	// fmt.Printf("Join filter: %s\nUnion filter: %s\nAfter filter: %s\n\n", n.filter, expressions.UnionFilters(n.left.Filter(), n.right.Filter()), expressions.FilterDifference(n.filter, union))
	n.filter = expressions.FilterDifference(n.filter, expressions.UnionFilters(n.left.Filter(), n.right.Filter()))
	n.strategy = &nestedLoopJoinStrategy{n: n}
}

func (n *joinNode) Filter() impls.Expression {
	return expressions.UnionFilters(n.filter, n.left.Filter(), n.right.Filter())
}

func (n *joinNode) Ordering() impls.OrderExpression {
	if n.strategy == nil {
		panic("No strategy set - optimization required before ordering can be determined")
	}

	return n.strategy.Ordering()
}

func (n *joinNode) SupportsMarkRestore() bool {
	return false
}

func (n *joinNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	if n.strategy == nil {
		panic("No strategy set - optimization required before scanning can be performed")
	}

	return n.strategy.Scanner(ctx)
}
