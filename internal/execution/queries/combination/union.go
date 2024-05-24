package combination

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution"
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/filter"
	"github.com/efritz/gostgres/internal/execution/queries/order"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
)

type unionNode struct {
	left     queries.Node
	right    queries.Node
	fields   []shared.Field
	distinct bool
}

var _ queries.Node = &unionNode{}

func NewUnion(left queries.Node, right queries.Node, distinct bool) (queries.Node, error) {
	leftFields := left.Fields()
	rightFields := right.Fields()

	if len(leftFields) != len(rightFields) {
		return nil, fmt.Errorf("unexpected union columns")
	}
	for i, leftField := range leftFields {
		if leftField.Type() != rightFields[i].Type() {
			// TODO - refine type if possible
			return nil, fmt.Errorf("unexpected union types")
		}
	}

	return &unionNode{
		left:     left,
		right:    right,
		fields:   leftFields,
		distinct: distinct,
	}, nil
}

func (n *unionNode) Name() string {
	return ""
}

func (n *unionNode) Fields() []shared.Field {
	return slices.Clone(n.fields)
}

func (n *unionNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("union")
	n.left.Serialize(w.Indent())
	w.WritefLine("with")
	n.right.Serialize(w.Indent())
}

func (n *unionNode) AddFilter(filterExpression expressions.Expression) {
	filter.LowerFilter(filterExpression, n.left, n.right)
}

func (n *unionNode) AddOrder(orderExpression expressions.OrderExpression) {
	order.LowerOrder(orderExpression, n.left, n.right)
}

func (n *unionNode) Optimize() {
	n.left.Optimize()
	n.right.Optimize()
}

func (n *unionNode) Filter() expressions.Expression {
	return expressions.FilterIntersection(n.left.Filter(), n.right.Filter())
}

func (n *unionNode) Ordering() expressions.OrderExpression { return nil }
func (n *unionNode) SupportsMarkRestore() bool             { return false }

func (n *unionNode) Scanner(ctx execution.Context) (scan.Scanner, error) {
	hash := map[string]struct{}{}
	mark := func(row shared.Row) bool {
		key := shared.HashSlice(row.Values)
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

	var rightScanner scan.Scanner

	return scan.ScannerFunc(func() (shared.Row, error) {
		for leftScanner != nil {
			row, err := leftScanner.Scan()
			if err != nil {
				if err == scan.ErrNoRows {
					leftScanner = nil
					continue
				}

				return shared.Row{}, err
			}

			if mark(row) {
				return row, nil
			}
		}

		if rightScanner == nil {
			rightScanner, err = n.right.Scanner(ctx)
			if err != nil {
				return shared.Row{}, err
			}
		}

		for {
			row, err := rightScanner.Scan()
			if err != nil {
				return shared.Row{}, err
			}

			if mark(row) {
				return shared.NewRow(n.Fields(), row.Values)
			}
		}
	}), nil
}
