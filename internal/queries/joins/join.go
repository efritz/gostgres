package joins

import (
	"fmt"
	"io"
	"slices"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/queries/filter"
	"github.com/efritz/gostgres/internal/queries/order"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
)

type joinNode struct {
	left     queries.Node
	right    queries.Node
	filter   expressions.Expression
	fields   []shared.Field
	strategy joinStrategy
}

var _ queries.Node = &joinNode{}

func NewJoin(left queries.Node, right queries.Node, condition expressions.Expression) queries.Node {
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

func (n *joinNode) Fields() []shared.Field {
	return slices.Clone(n.fields)
}

func (n *joinNode) Serialize(w io.Writer, indentationLevel int) {
	indentation := serialization.Indent(indentationLevel)
	io.WriteString(w, fmt.Sprintf("%sjoin using %s\n", indentation, n.strategy.Name()))
	n.left.Serialize(w, indentationLevel+1)
	io.WriteString(w, fmt.Sprintf("%swith\n", indentation))
	n.right.Serialize(w, indentationLevel+1)

	if n.filter != nil {
		io.WriteString(w, fmt.Sprintf("%son %s\n", indentation, n.filter))
	}
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

func bindsAllFields(n queries.Node, expr expressions.Expression) bool {
	for _, field := range expr.Fields() {
		if _, err := shared.FindMatchingFieldIndex(field, n.Fields()); err != nil {
			return false
		}
	}

	return true
}

func (n *joinNode) AddFilter(filterExpression expressions.Expression) {
	n.filter = expressions.UnionFilters(n.filter, filterExpression)
}

func (n *joinNode) AddOrder(orderExpression expressions.OrderExpression) {
	order.LowerOrder(orderExpression, n.left, n.right)
}

func (n *joinNode) Filter() expressions.Expression {
	return expressions.UnionFilters(n.filter, n.left.Filter(), n.right.Filter())
}

func (n *joinNode) Ordering() expressions.OrderExpression {
	if n.strategy == nil {
		panic("No strategy set - optimization required before ordering can be determined")
	}

	return n.strategy.Ordering()
}

func (n *joinNode) SupportsMarkRestore() bool {
	return false
}

func (n *joinNode) Scanner(ctx queries.Context) (scan.Scanner, error) {
	if n.strategy == nil {
		panic("No strategy set - optimization required before scanning can be performed")
	}

	return n.strategy.Scanner(ctx)
}
