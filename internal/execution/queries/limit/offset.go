package limit

import (
	"github.com/efritz/gostgres/internal/execution/engine/serialization"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
)

type offsetNode struct {
	queries.Node
	offset int
}

var _ queries.Node = &offsetNode{}

func NewOffset(node queries.Node, offset int) queries.Node {
	return &offsetNode{
		Node:   node,
		offset: offset,
	}
}

func (n *offsetNode) Serialize(w serialization.IndentWriter) {
	if n.offset == 0 {
		n.Node.Serialize(w)
	} else {
		w.WritefLine("offset %d", n.offset)
		n.Node.Serialize(w.Indent())
	}
}

func (n *offsetNode) AddFilter(filter impls.Expression)    {}
func (n *offsetNode) AddOrder(order impls.OrderExpression) {}
func (n *offsetNode) Optimize()                            { n.Node.Optimize() }
func (n *offsetNode) Filter() impls.Expression             { return n.Node.Filter() }
func (n *offsetNode) Ordering() impls.OrderExpression      { return n.Node.Ordering() }
func (n *offsetNode) SupportsMarkRestore() bool            { return false }

func (n *offsetNode) Scanner(ctx impls.Context) (scan.Scanner, error) {
	ctx.Log("Building Offset scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	offset := n.offset
	if offset == 0 {
		return scanner, nil
	}

	return scan.ScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Offset")

		for {
			row, err := scanner.Scan()
			if err != nil {
				return rows.Row{}, err
			}

			offset--
			if offset >= 0 {
				continue
			}

			return row, err
		}
	}), nil
}
