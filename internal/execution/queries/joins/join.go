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

func (n *joinNode) AddFilter(filterExpression impls.Expression) {
	n.filter = expressions.UnionFilters(n.filter, filterExpression)
}

func (n *joinNode) AddOrder(orderExpression impls.OrderExpression) {
	order.LowerOrder(orderExpression, n.left, n.right)
}

func (n *joinNode) Optimize() {
	if n.filter != nil {
		n.filter = n.filter.Fold()
		filter.LowerFilter(n.filter, n.left, n.right)
	}

	n.left.Optimize()
	n.right.Optimize()
	n.filter = expressions.FilterDifference(n.filter, expressions.UnionFilters(n.left.Filter(), n.right.Filter()))
	n.strategy = selectJoinStrategy(n)
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

func (n *joinNode) Scanner(ctx impls.Context) (scan.RowScanner, error) {
	if n.strategy == nil {
		panic("No strategy set - optimization required before scanning can be performed")
	}

	return n.strategy.Scanner(ctx)
}
