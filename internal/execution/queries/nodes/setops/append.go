package setops

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type appendNode struct {
	left   nodes.Node
	right  nodes.Node
	fields []fields.Field
}

func NewAppend(left nodes.Node, right nodes.Node, fields []fields.Field) nodes.Node {
	return &appendNode{
		left:   left,
		right:  right,
		fields: fields,
	}
}

func (n *appendNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("append")
	n.left.Serialize(w.Indent())
	w.WritefLine("and")
	n.right.Serialize(w.Indent())
}

func (n *appendNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Append scanner")

	leftScanner, err := n.left.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	rightScanner, err := n.right.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Append")

		for leftScanner != nil {
			row, err := leftScanner.Scan()
			if err != nil {
				if err == scan.ErrNoRows {
					leftScanner = nil
					continue
				}

				return rows.Row{}, err
			}

			return row, nil
		}

		row, err := rightScanner.Scan()
		if err != nil {
			return rows.Row{}, err
		}

		return rows.NewRow(n.fields, row.Values)
	}), nil
}
