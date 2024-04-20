package joins

import (
	"fmt"
	"io"
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/scan"
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
	return copyFields(n.fields)
}

func (n *joinNode) Serialize(w io.Writer, indentationLevel int) {
	indentation := indent(indentationLevel)
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
		lowerFilter(n.filter, n.left, n.right)
	}

	n.left.Optimize()
	n.right.Optimize()
	n.filter = filterDifference(n.filter, unionFilters(n.left.Filter(), n.right.Filter()))
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

func (n *joinNode) AddFilter(filter expressions.Expression) {
	n.filter = unionFilters(n.filter, filter)
}

func (n *joinNode) AddOrder(order expressions.OrderExpression) {
	lowerOrder(order, n.left, n.right)
}

func (n *joinNode) Filter() expressions.Expression {
	return unionFilters(n.filter, n.left.Filter(), n.right.Filter())
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

func (n *joinNode) Scanner(ctx scan.ScanContext) (scan.Scanner, error) {
	if n.strategy == nil {
		panic("No strategy set - optimization required before scanning can be performed")
	}

	return n.strategy.Scanner(ctx)
}

// TODO - deduplicate

func copyFields(fields []shared.Field) []shared.Field {
	c := make([]shared.Field, len(fields))
	copy(c, fields)
	return c
}

func indent(level int) string {
	return strings.Repeat(" ", level*4)
}

func filterDifference(filter, childFilter expressions.Expression) expressions.Expression {
	return combineFilters(filter, childFilter, func(conjunctions, childConjunctions []expressions.Expression) {
		for i, f1 := range conjunctions {
			for _, f2 := range childConjunctions {
				if f1.Equal(f2) {
					conjunctions[i] = nil
					break
				}
			}
		}
	})
}

func combineFilters(filter, childFilter expressions.Expression, filterConjunctions func(conjunctions, childConjunctions []expressions.Expression)) expressions.Expression {
	if filter == nil {
		return nil
	}
	if childFilter == nil {
		return filter
	}

	conjunctions := filter.Conjunctions()
	filterConjunctions(conjunctions, childFilter.Conjunctions())
	return unionFilters(conjunctions...)
}

func unionFilters(filters ...expressions.Expression) expressions.Expression {
	var conjunctions []expressions.Expression
	for _, expression := range filters {
		if expression == nil {
			continue
		}

		conjunctions = append(conjunctions, expression.Conjunctions()...)
	}
	if len(conjunctions) == 0 {
		return nil
	}

	for i, c1 := range conjunctions {
		for j, c2 := range conjunctions {
			if c1 == nil || c2 == nil || j <= i {
				continue
			}

			if c1.Equal(c2) {
				conjunctions[j] = nil
			}
		}
	}

	filter := conjunctions[0]
	for _, conjunction := range conjunctions[1:] {
		if conjunction == nil {
			continue
		}

		filter = expressions.NewAnd(filter, conjunction)
	}

	return filter
}

func lowerOrder(order expressions.OrderExpression, nodes ...queries.Node) {
	orderExpressions := order.Expressions()

	for _, node := range nodes {
		filteredExpressions := make([]expressions.ExpressionWithDirection, 0, len(orderExpressions))
	exprLoop:
		for _, expression := range orderExpressions {
			for _, field := range expression.Expression.Fields() {
				if _, err := shared.FindMatchingFieldIndex(field, node.Fields()); err != nil {
					continue exprLoop
				}
			}

			filteredExpressions = append(filteredExpressions, expression)
		}

		if len(filteredExpressions) != 0 {
			node.AddOrder(expressions.NewOrderExpression(filteredExpressions))
		}
	}
}

func lowerFilter(filter expressions.Expression, nodes ...queries.Node) {
	for _, expression := range filter.Conjunctions() {
		missing := make([]bool, len(nodes))
		for _, field := range expression.Fields() {
			for i, node := range nodes {
				if _, err := shared.FindMatchingFieldIndex(field, node.Fields()); err != nil {
					missing[i] = true
				}
			}
		}

		for i, missing := range missing {
			if !missing {
				nodes[i].AddFilter(expression)
			}
		}
	}
}
