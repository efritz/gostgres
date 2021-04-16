package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
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
	}

	n.Node.Optimize()
}

func (n *orderNode) PushDownFilter(filter expressions.Expression) bool {
	return n.Node.PushDownFilter(filter)
}

func (n *orderNode) Scan(visitor VisitorFunc) error {
	if n.order == nil {
		return n.Node.Scan(visitor)
	}

	rows, err := ScanRows(n.Node)
	if err != nil {
		return err
	}

	indexes, err := findIndexIterationOrder(n.order, rows)
	if err != nil {
		return err
	}

	for _, i := range indexes {
		if ok, err := visitor(rows.Row(i)); err != nil {
			return err
		} else if !ok {
			break
		}
	}

	return nil
}
