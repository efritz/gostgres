package limit

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type logicalLimitNode struct {
	queries.LogicalNode
	limit int
}

var _ queries.LogicalNode = &logicalLimitNode{}

func NewLimit(node queries.LogicalNode, limit int) queries.LogicalNode {
	return &logicalLimitNode{
		LogicalNode: node,
		limit:       limit,
	}
}

func (n *logicalLimitNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {} // boundary
func (n *logicalLimitNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {} // boundary
func (n *logicalLimitNode) SupportsMarkRestore() bool                                           { return false }

func (n *logicalLimitNode) Build() queries.Node {
	return &limitNode{
		Node:  n.LogicalNode.Build(),
		limit: n.limit,
	}
}

//
//

type limitNode struct {
	queries.Node
	limit int
}

var _ queries.Node = &limitNode{}

func (n *limitNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("limit %d", n.limit)
	n.Node.Serialize(w.Indent())
}

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
