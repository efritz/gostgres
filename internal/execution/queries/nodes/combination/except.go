package combination

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/utils"
)

type exceptNode struct {
	left     nodes.Node
	right    nodes.Node
	fields   []fields.Field
	distinct bool
}

func NewExcept(left, right nodes.Node, fields []fields.Field, distinct bool) nodes.Node {
	return &exceptNode{
		left:     left,
		right:    right,
		fields:   fields,
		distinct: distinct,
	}
}

func (n *exceptNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("except")
	n.left.Serialize(w.Indent())
	w.WritefLine("with")
	n.right.Serialize(w.Indent())
}

func (n *exceptNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Except scanner")

	leftScanner, err := n.left.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	rightScanner, err := n.right.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	rowsToIgnore := map[string]rowWithCount{}
	if err := scan.VisitRows(rightScanner, hashVisitor(rowsToIgnore)); err != nil {
		return nil, err
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Except")

		for {
			row, err := leftScanner.Scan()
			if err != nil {
				return rows.Row{}, err
			}

			key := utils.HashSlice(row.Values)
			rc, ok := rowsToIgnore[key]
			if !ok {
				if n.distinct {
					// This is an "EXCEPT DISTINCT" query, and we need to track the set of rows
					// we emit so we don't emit a duplicate later. We can reuse the same list.
					// The value of the count here is not relevant, as we only adjust and check
					// this value for "EXCEPT ALL" queries.
					rowsToIgnore[key] = rowWithCount{Row: row, Count: 1}
				}

				// The row wasn't in the ignore list, so can emit it.
				return row, nil
			}

			if !n.distinct {
				// For "EXCEPT ALL" queries, we only want to ignore the same number of equivalent
				// rows from the left relation that exist in the right relation. We'll adjust the
				// count of the rows in the ignore list, and remove the row completely once zero.

				if rc.Count > 1 {
					rowsToIgnore[key] = rowWithCount{Row: row, Count: rc.Count - 1}
				} else {
					delete(rowsToIgnore, key)
				}
			}
		}
	}), nil
}
