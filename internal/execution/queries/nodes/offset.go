package nodes

import (
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type offsetNode struct {
	Node
	offset int
}

func NewOffset(node Node, offset int) Node {
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

func (n *offsetNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Offset scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	offset := n.offset
	if offset == 0 {
		return scanner, nil
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
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
