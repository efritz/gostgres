package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type limitNode struct {
	Node
	limit int
}

var _ Node = &limitNode{}

func NewLimit(node Node, limit int) Node {
	return &limitNode{
		Node:  node,
		limit: limit,
	}
}

func (n *limitNode) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%slimit %d\n", indent(indentationLevel), n.limit))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *limitNode) Optimize() {
	n.Node.Optimize()
}

func (n *limitNode) AddFilter(filter expressions.Expression) {
	// filter boundary: do not recurse
}

func (n *limitNode) AddOrder(order OrderExpression) {
	// filter boundary: do not recurse
}

func (n *limitNode) Filter() expressions.Expression {
	return n.Node.Filter()
}

func (n *limitNode) Ordering() OrderExpression {
	return n.Node.Ordering()
}

func (n *limitNode) SupportsMarkRestore() bool {
	return false
}

func (n *limitNode) Scanner(ctx ScanContext) (Scanner, error) {
	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	remaining := n.limit

	return ScannerFunc(func() (shared.Row, error) {
		if remaining <= 0 {
			return shared.Row{}, ErrNoRows
		}

		remaining--
		return scanner.Scan()
	}), nil
}
