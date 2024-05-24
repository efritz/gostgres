package combination

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/filter"
	"github.com/efritz/gostgres/internal/execution/queries/order"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
)

type combinationNode struct {
	left             queries.Node
	right            queries.Node
	fields           []shared.Field
	groupedRowFilter groupedRowFilterFunc
	distinct         bool
}

var _ queries.Node = &combinationNode{}

type sourcedRow struct {
	index int
	row   shared.Row
}

type groupedRowFilterFunc func(rows []sourcedRow) bool

func NewIntersect(left queries.Node, right queries.Node, distinct bool) (queries.Node, error) {
	return newCombination(left, right, intersectFilter, distinct)
}

func NewExcept(left queries.Node, right queries.Node, distinct bool) (queries.Node, error) {
	return newCombination(left, right, exceptFilter, distinct)
}

func newCombination(left queries.Node, right queries.Node, groupedRowFilter groupedRowFilterFunc, distinct bool) (queries.Node, error) {
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

	return &combinationNode{
		left:             left,
		right:            right,
		fields:           leftFields,
		groupedRowFilter: groupedRowFilter,
		distinct:         distinct,
	}, nil
}

func (n *combinationNode) Name() string {
	return ""
}

func (n *combinationNode) Fields() []shared.Field {
	return slices.Clone(n.fields)
}

func (n *combinationNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("combination")
	n.left.Serialize(w.Indent())
	w.WritefLine("with")
	n.right.Serialize(w.Indent())
}

func (n *combinationNode) AddFilter(filterExpression expressions.Expression) {
	filter.LowerFilter(filterExpression, n.left, n.right)
}

func (n *combinationNode) AddOrder(orderExpression expressions.OrderExpression) {
	order.LowerOrder(orderExpression, n.left, n.right)
}

func (n *combinationNode) Optimize() {
	n.left.Optimize()
	n.right.Optimize()
}

func (n *combinationNode) Filter() expressions.Expression {
	return n.left.Filter()
}

func (n *combinationNode) Ordering() expressions.OrderExpression { return nil }
func (n *combinationNode) SupportsMarkRestore() bool             { return false }

func (n *combinationNode) Scanner(ctx queries.Context) (scan.Scanner, error) {
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

	return scan.ScannerFunc(func() (shared.Row, error) {
	outer:
		for {
			if len(selection) > 0 {
				row := selection[0]
				selection = selection[1:]
				return shared.NewRow(n.Fields(), row.row.Values)
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

		return shared.Row{}, scan.ErrNoRows
	}), nil
}

func hashVisitor(hash map[string][]sourcedRow, index int) scan.VisitorFunc {
	return func(row shared.Row) (bool, error) {
		key := shared.HashSlice(row.Values)

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
