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

func (n *offsetNode) SupportsMarkRestore() bool {
	return false
}

func (n *offsetNode) Scanner(ctx ScanContext) (Scanner, error) {
	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	offset := n.offset
	if offset == 0 {
		return scanner, nil
	}

	return ScannerFunc(func() (shared.Row, error) {
		for {
			row, err := scanner.Scan()
			if err != nil {
				return shared.Row{}, err
			}

			offset--
			if offset >= 0 {
				continue
			}

			return row, err
		}
	}), nil
}
