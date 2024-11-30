package nodes

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/utils"
)

type logicalCombinationNode struct {
	left             LogicalNode
	right            LogicalNode
	fields           []fields.Field
	groupedRowFilter groupedRowFilterFunc
	distinct         bool
}

var _ LogicalNode = &logicalCombinationNode{}

type sourcedRow struct {
	index int
	row   rows.Row
}

type groupedRowFilterFunc func(rows []sourcedRow) bool

func NewIntersect(left LogicalNode, right LogicalNode, distinct bool) (LogicalNode, error) {
	return newCombination(left, right, intersectFilter, distinct)
}

func NewExcept(left LogicalNode, right LogicalNode, distinct bool) (LogicalNode, error) {
	return newCombination(left, right, exceptFilter, distinct)
}

func newCombination(left LogicalNode, right LogicalNode, groupedRowFilter groupedRowFilterFunc, distinct bool) (LogicalNode, error) {
	leftFields := left.Fields()
	rightFields := right.Fields()

	if len(leftFields) != len(rightFields) {
		return nil, fmt.Errorf("unexpected combination columns")
	}
	for i, leftField := range leftFields {
		if leftField.Type() != rightFields[i].Type() {
			// TODO - refine type if possible
			return nil, fmt.Errorf("unexpected combination types")
		}
	}

	return &logicalCombinationNode{
		left:             left,
		right:            right,
		fields:           leftFields,
		groupedRowFilter: groupedRowFilter,
		distinct:         distinct,
	}, nil
}

func (n *logicalCombinationNode) Name() string {
	return ""
}

func (n *logicalCombinationNode) Fields() []fields.Field {
	return slices.Clone(n.fields)
}

func (n *logicalCombinationNode) AddFilter(ctx impls.OptimizationContext, filterExpression impls.Expression) {
	lowerFilter(ctx, filterExpression, n.left, n.right)
}

func (n *logicalCombinationNode) AddOrder(ctx impls.OptimizationContext, orderExpression impls.OrderExpression) {
	lowerOrder(ctx, orderExpression, n.left, n.right)
}

func (n *logicalCombinationNode) Optimize(ctx impls.OptimizationContext) {
	n.left.Optimize(ctx)
	n.right.Optimize(ctx)
}

func (n *logicalCombinationNode) Filter() impls.Expression {
	return n.left.Filter()
}

func (n *logicalCombinationNode) Ordering() impls.OrderExpression { return nil }
func (n *logicalCombinationNode) SupportsMarkRestore() bool       { return false }

func (n *logicalCombinationNode) Build() Node {
	return &combinationNode{
		left:             n.left.Build(),
		right:            n.right.Build(),
		fields:           n.fields,
		groupedRowFilter: n.groupedRowFilter,
		distinct:         n.distinct,
	}
}

//
//

type combinationNode struct {
	left             Node
	right            Node
	fields           []fields.Field
	groupedRowFilter groupedRowFilterFunc
	distinct         bool
}

var _ Node = &combinationNode{}

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

	hash := map[string][]sourcedRow{}
	if err := scan.VisitRows(leftScanner, hashVisitor(hash, 0)); err != nil {
		return nil, err
	}
	if err := scan.VisitRows(rightScanner, hashVisitor(hash, 1)); err != nil {
		return nil, err
	}

	var selection []sourcedRow

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Combination")

	outer:
		for {
			if len(selection) > 0 {
				row := selection[0]
				selection = selection[1:]
				return rows.NewRow(n.fields, row.row.Values)
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

func hashVisitor(hash map[string][]sourcedRow, index int) scan.VisitorFunc {
	return func(row rows.Row) (bool, error) {
		key := utils.HashSlice(row.Values)

		hash[key] = append(hash[key], sourcedRow{
			index: index,
			row:   row,
		})

		return true, nil
	}
}

func intersectFilter(rows []sourcedRow) bool {
	indexes := map[int]bool{}
	for _, r := range rows {
		indexes[r.index] = true
	}

	if !indexes[0] || !indexes[1] {
		return false
	}

	return true
}

func exceptFilter(rows []sourcedRow) bool {
	for _, r := range rows {
		if r.index == 1 {
			return false
		}
	}

	return true
}
