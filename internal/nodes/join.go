package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type joinNode struct {
	left     Node
	right    Node
	filter   expressions.Expression
	fields   []shared.Field
	strategy joinStrategy
}

type joinStrategy interface {
	Name() string
	Ordering() OrderExpression
	Scan(visitor VisitorFunc) error
}

var _ Node = &joinNode{}

func NewJoin(left Node, right Node, condition expressions.Expression) Node {
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
	n.strategy = n.selectStrategy()
}

func (n *joinNode) selectStrategy() joinStrategy {
	comparisonType, left, right := n.decomposeFilter()
	if comparisonType == expressions.ComparisonTypeEquals {
		return &hashJoinStrategy{
			n:     n,
			left:  left,
			right: right,
		}
	}

	// TODO - determine condition
	if false {
		return &mergeJoinStrategy{
			n:              n,
			left:           left,
			right:          right,
			comparisonType: comparisonType,
		}
	}

	return &nestedLoopJoinStrategy{n: n}
}

func (n *joinNode) decomposeFilter() (_ expressions.ComparisonType, left, right expressions.Expression) {
	if comparisonType, left, right := expressions.IsComparison(n.filter); comparisonType != expressions.ComparisonTypeUnknown {
		if bindsAllFields(n.left, left) && bindsAllFields(n.right, right) {
			return comparisonType, left, right
		}

		if bindsAllFields(n.left, right) && bindsAllFields(n.right, left) {
			return comparisonType.Flip(), right, left
		}
	}

	return expressions.ComparisonTypeUnknown, nil, nil
}

func bindsAllFields(n Node, expr expressions.Expression) bool {
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

func (n *joinNode) AddOrder(order OrderExpression) {
	lowerOrder(order, n.left, n.right)
}

func (n *joinNode) Filter() expressions.Expression {
	return unionFilters(n.filter, n.left.Filter(), n.right.Filter())
}

func (n *joinNode) Ordering() OrderExpression {
	if n.strategy == nil {
		panic("No strategy set - optimization required before ordering can be determined")
	}

	return n.strategy.Ordering()
}

func (n *joinNode) Scan(visitor VisitorFunc) error {
	if n.strategy == nil {
		panic("No strategy set - optimization required before scanning can be performed")
	}

	return n.strategy.Scan(visitor)
}
