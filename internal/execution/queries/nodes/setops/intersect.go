package setops

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/utils"
)

type intersectNode struct {
	left     nodes.Node
	right    nodes.Node
	fields   []fields.Field
	distinct bool
}

func NewIntersect(left, right nodes.Node, fields []fields.Field, distinct bool) nodes.Node {
	return &intersectNode{
		left:     left,
		right:    right,
		fields:   fields,
		distinct: distinct,
	}
}

func (n *intersectNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("intersect")
	n.left.Serialize(w.Indent())
	w.WritefLine("with")
	n.right.Serialize(w.Indent())
}

func (n *intersectNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Intersect scanner")

	leftScanner, err := n.left.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	// We assume the right relation is smaller
	rightScanner, err := n.right.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	rowsPresent := map[string]rowWithCount{}
	if err := scan.VisitRows(rightScanner, hashVisitor(rowsPresent)); err != nil {
		return nil, err
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Intersect")

		for {
			row, err := leftScanner.Scan()
			if err != nil {
				return rows.Row{}, err
			}

			key := utils.HashSlice(row.Values)
			rc, ok := rowsPresent[key]
			if !ok {
				// No matching row found in right relation
				continue
			}

			// For "INTERSECT DISTINCT" queries, we want to emit the first row from the left
			// relation that matches a row in the right relation. After we do so, we remove
			// the matching row from the present set so that future mathching rows don't emit
			// a duplicate.
			//
			// For "INTERSECT ALL" queries, we care about multiplicity. Every time we find a
			// match, we'll decrease its count by one. Once it hits zero we'll remove it from
			// the present set.

			if !n.distinct && rc.Count > 1 {
				rowsPresent[key] = rowWithCount{Row: rc.Row, Count: rc.Count - 1}
			} else {
				delete(rowsPresent, key)
			}

			// There was a matching row in the right relation, so we can emit it.
			return row, nil
		}
	}), nil
}

//
//

type rowWithCount struct {
	Row   rows.Row
	Count int
}

func hashVisitor(hash map[string]rowWithCount) scan.VisitorFunc {
	return func(row rows.Row) (bool, error) {
		key := utils.HashSlice(row.Values)

		if rc, ok := hash[key]; ok {
			hash[key] = rowWithCount{Row: row, Count: rc.Count + 1}
		} else {
			hash[key] = rowWithCount{Row: row, Count: 1}
		}

		return true, nil
	}
}
