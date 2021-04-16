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

func (n *offsetNode) PushDownFilter(filter expressions.Expression) bool {
	// filter boundary
	return false
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
