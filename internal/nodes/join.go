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
	// TODO - expand to handle (a == b AND c == d)
	if expressions.IsEquality(n.filter) {
		if fields := n.filter.Fields(); len(fields) == 2 {
			_, err1 := shared.FindMatchingFieldIndex(fields[0], n.left.Fields())
			_, err2 := shared.FindMatchingFieldIndex(fields[1], n.right.Fields())
			_, err3 := shared.FindMatchingFieldIndex(fields[1], n.left.Fields())
			_, err4 := shared.FindMatchingFieldIndex(fields[0], n.right.Fields())

			if (err1 == nil && err2 == nil) && (err3 != nil && err4 != nil) {
				return &hashJoinStrategy{
					n:          n,
					leftField:  fields[0],
					rightField: fields[1],
				}
			}

			if (err1 != nil && err2 != nil) && (err3 == nil && err4 == nil) {
				return &hashJoinStrategy{
					n:          n,
					leftField:  fields[1],
					rightField: fields[0],
				}
			}
		}
	}

	// TODO - attempt merge join strategy when applicable
	if false {
		return &mergeJoinStrategy{n: n}
	}

	return &nestedLoopJoinStrategy{n: n}
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
