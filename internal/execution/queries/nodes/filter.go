package nodes

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/types"
)

type filterNode struct {
	Node
	filter impls.Expression
}

func NewFilter(node Node, filter impls.Expression) Node {
	return &filterNode{
		Node:   node,
		filter: filter,
	}
}

func (n *filterNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("filter by %s", n.filter)
	n.Node.Serialize(w.Indent())
}

func (n *filterNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Filter Node scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Filter")

		for {
			row, err := scanner.Scan()
			if err != nil {
				return rows.Row{}, err
			}

			if ok, err := types.ValueAs[bool](queries.Evaluate(ctx, n.filter, row)); err != nil {
				return rows.Row{}, err
			} else if ok == nil || !*ok {
				continue
			}

			return row, nil
		}
	}), nil
}
