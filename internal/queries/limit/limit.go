package limit

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
)

type limitNode struct {
	queries.Node
	limit int
}

var _ queries.Node = &limitNode{}

func NewLimit(node queries.Node, limit int) queries.Node {
	return &limitNode{
		Node:  node,
		limit: limit,
	}
}

func (n *limitNode) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%slimit %d\n", serialization.Indent(indentationLevel), n.limit))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *limitNode) Optimize() {
	n.Node.Optimize()
}

func (n *limitNode) AddFilter(filter expressions.Expression) {
	// filter boundary: do not recurse
}

func (n *limitNode) AddOrder(order expressions.OrderExpression) {
	// filter boundary: do not recurse
}

func (n *limitNode) Filter() expressions.Expression {
	return n.Node.Filter()
}

func (n *limitNode) Ordering() expressions.OrderExpression {
	return n.Node.Ordering()
}

func (n *limitNode) SupportsMarkRestore() bool {
	return false
}

func (n *limitNode) Scanner(ctx queries.Context) (scan.Scanner, error) {
	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	remaining := n.limit

	return scan.ScannerFunc(func() (shared.Row, error) {
		if remaining <= 0 {
			return shared.Row{}, scan.ErrNoRows
		}

		remaining--
		return scanner.Scan()
	}), nil
}
