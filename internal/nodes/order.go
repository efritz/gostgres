package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type orderNode struct {
	Node
	order OrderExpression
}

var _ Node = &orderNode{}

func NewOrder(node Node, order OrderExpression) Node {
	return &orderNode{
		Node:  node,
		order: order,
	}
}

func (n *orderNode) Serialize(w io.Writer, indentationLevel int) {
	if n.order == nil {
		n.Node.Serialize(w, indentationLevel)
		return
	}

	io.WriteString(w, fmt.Sprintf("%sorder by %s\n", indent(indentationLevel), n.order))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *orderNode) Optimize() {
	if n.order != nil {
		n.order = n.order.Fold()
		n.Node.AddOrder(n.order)
	}

	n.Node.Optimize()

	if n.order == nil {
		return
	}

	childOrdering := n.Node.Ordering()
	if childOrdering == nil {
		return
	}

	expressions := n.order.Expressions()
	childExpressions := childOrdering.Expressions()
	if len(childExpressions) < len(expressions) {
		return
	}

	for i, expression := range expressions {
		if expression.Reverse != childExpressions[i].Reverse {
			return
		}

		if !expression.Expression.Equal(childExpressions[i].Expression) {
			return
		}
	}

	n.order = nil
}

func (n *orderNode) AddFilter(filter expressions.Expression) {
	n.Node.AddFilter(filter)
}

func (n *orderNode) AddOrder(order OrderExpression) {
	// We are nested in a parent sort and un-separated by an ordering boundary
	// (such as limit or offset). We'll ignore our old sort criteria and adopt
	// our parent since the ordering of rows at this point in the query should
	// not have an effect on the result.
	n.order = order
}

func (n *orderNode) Ordering() OrderExpression {
	if n.order == nil {
		return n.Node.Ordering()
	}

	return n.order
}

func (n *orderNode) Scanner() (Scanner, error) {
	scanner, err := n.Node.Scanner()
	if err != nil {
		return nil, err
	}

	if n.order == nil {
		return scanner, nil
	}

	rows, err := ScanRows(n.Node)
	if err != nil {
		return nil, err
	}

	indexes, err := findIndexIterationOrder(n.order, rows)
	if err != nil {
		return nil, err
	}

	i := 0

	return ScannerFunc(func() (shared.Row, error) {
		if i < len(indexes) {
			row := rows.Row(indexes[i])
			i++
			return row, nil
		}

		return shared.Row{}, ErrNoRows
	}), nil
}
