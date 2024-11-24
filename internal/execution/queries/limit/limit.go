package limit

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
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

func (n *limitNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {}
func (n *limitNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {}
func (n *limitNode) Optimize(ctx impls.OptimizationContext)                              { n.Node.Optimize(ctx) }
func (n *limitNode) Filter() impls.Expression                                            { return n.Node.Filter() }
func (n *limitNode) Ordering() impls.OrderExpression                                     { return n.Node.Ordering() }
func (n *limitNode) SupportsMarkRestore() bool                                           { return false }

func (n *limitNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Limit scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	remaining := n.limit

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Limit")

		if remaining <= 0 {
			return rows.Row{}, scan.ErrNoRows
		}

		remaining--
		return scanner.Scan()
	}), nil
}
