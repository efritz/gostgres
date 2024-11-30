package nodes

import (
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/utils"
)

type combinationNode struct {
	left             Node
	right            Node
	fields           []fields.Field
	groupedRowFilter GroupedRowFilterFunc
	distinct         bool
}

type SourcedRow struct {
	Index int
	Row   rows.Row
}

type GroupedRowFilterFunc func(rows []SourcedRow) bool

func NewCombination(left, right Node, fields []fields.Field, groupedRowFilter GroupedRowFilterFunc, distinct bool) Node {
	return &combinationNode{
		left:             left,
		right:            right,
		fields:           fields,
		groupedRowFilter: groupedRowFilter,
		distinct:         distinct,
	}
}

func (n *combinationNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("combination")
	n.left.Serialize(w.Indent())
	w.WritefLine("with")
	n.right.Serialize(w.Indent())
}

func (n *combinationNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Combination scanner")

	leftScanner, err := n.left.Scanner(ctx)
	if err != nil {
		return nil, err
	}
	rightScanner, err := n.right.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	hash := map[string][]SourcedRow{}
	if err := scan.VisitRows(leftScanner, hashVisitor(hash, 0)); err != nil {
		return nil, err
	}
	if err := scan.VisitRows(rightScanner, hashVisitor(hash, 1)); err != nil {
		return nil, err
	}

	var selection []SourcedRow

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Combination")

	outer:
		for {
			if len(selection) > 0 {
				row := selection[0]
				selection = selection[1:]
				return rows.NewRow(n.fields, row.Row.Values)
			}

			for key, rows := range hash {
				if n.groupedRowFilter(rows) {
					if n.distinct {
						selection = rows[:1]
					} else {
						selection = rows
					}
				}

				delete(hash, key)
				continue outer
			}

			break
		}

		return rows.Row{}, scan.ErrNoRows
	}), nil
}

func hashVisitor(hash map[string][]SourcedRow, index int) scan.VisitorFunc {
	return func(row rows.Row) (bool, error) {
		key := utils.HashSlice(row.Values)

		hash[key] = append(hash[key], SourcedRow{
			Index: index,
			Row:   row,
		})

		return true, nil
	}
}
