package nodes

import (
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/utils"
)

type unionNode struct {
	left     Node
	right    Node
	fields   []fields.Field
	distinct bool
}

func NewUnion(left Node, right Node, fields []fields.Field, distinct bool) Node {
	return &unionNode{
		left:     left,
		right:    right,
		fields:   fields,
		distinct: distinct,
	}
}

func (n *unionNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("union")
	n.left.Serialize(w.Indent())
	w.WritefLine("with")
	n.right.Serialize(w.Indent())
}

func (n *unionNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Union scanner")

	hash := map[string]struct{}{}
	mark := func(row rows.Row) bool {
		key := utils.HashSlice(row.Values)
		if _, ok := hash[key]; ok {
			return !n.distinct
		}

		hash[key] = struct{}{}
		return true
	}

	leftScanner, err := n.left.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	var rightScanner scan.RowScanner

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Union")

		for leftScanner != nil {
			row, err := leftScanner.Scan()
			if err != nil {
				if err == scan.ErrNoRows {
					leftScanner = nil
					continue
				}

				return rows.Row{}, err
			}

			if mark(row) {
				return row, nil
			}
		}

		if rightScanner == nil {
			rightScanner, err = n.right.Scanner(ctx)
			if err != nil {
				return rows.Row{}, err
			}
		}

		for {
			row, err := rightScanner.Scan()
			if err != nil {
				return rows.Row{}, err
			}

			if mark(row) {
				return rows.NewRow(n.fields, row.Values)
			}
		}
	}), nil
}
