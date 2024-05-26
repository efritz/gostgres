package limit

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
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

func (n *limitNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("limit %d", n.limit)
	n.Node.Serialize(w.Indent())
}

func (n *limitNode) AddFilter(filter types.Expression)    {}
func (n *limitNode) AddOrder(order types.OrderExpression) {}
func (n *limitNode) Optimize()                            { n.Node.Optimize() }
func (n *limitNode) Filter() types.Expression             { return n.Node.Filter() }
func (n *limitNode) Ordering() types.OrderExpression      { return n.Node.Ordering() }
func (n *limitNode) SupportsMarkRestore() bool            { return false }

func (n *limitNode) Scanner(ctx types.Context) (scan.Scanner, error) {
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
