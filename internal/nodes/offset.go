package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type offsetNode struct {
	Node
	offset int
}

var _ Node = &offsetNode{}

func NewOffset(node Node, offset int) Node {
	return &offsetNode{
		Node:   node,
		offset: offset,
	}
}

func (n *offsetNode) Serialize(w io.Writer, indentationLevel int) {
	if n.offset == 0 {
		n.Node.Serialize(w, indentationLevel)
		return
	}

	io.WriteString(w, fmt.Sprintf("%soffset %d\n", indent(indentationLevel), n.offset))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *offsetNode) Optimize() {
	n.Node.Optimize()
}

func (n *offsetNode) AddFilter(filter expressions.Expression) {
	// filter boundary: do not recurse
}

func (n *offsetNode) AddOrder(order OrderExpression) {
	// order boundary: do not recurse
}

func (n *offsetNode) Filter() expressions.Expression {
	return n.Node.Filter()
}

func (n *offsetNode) Ordering() OrderExpression {
	return n.Node.Ordering()
}

func (n *offsetNode) Scan(visitor VisitorFunc) error {
	if n.offset == 0 {
		return n.Node.Scan(visitor)
	}

	offset := n.offset
	return n.Node.Scan(func(row shared.Row) (bool, error) {
		offset--
		if offset >= 0 {
			return true, nil
		}

		return visitor(row)
	})
}
