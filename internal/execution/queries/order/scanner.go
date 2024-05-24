package order

import (
	"fmt"
	"sort"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/shared"
)

type orderScanner struct {
	rows    shared.Rows
	indexes []int
	next    int
	mark    int
}

func NewOrderScanner(ctx queries.Context, scanner scan.Scanner, fields []shared.Field, order expressions.OrderExpression) (scan.Scanner, error) {
	rows, err := shared.NewRows(fields)
	if err != nil {
		return nil, err
	}

	rows, err = scan.ScanIntoRows(scanner, rows)
	if err != nil {
		return nil, err
	}

	indexes, err := findIndexIterationOrder(ctx, order, rows)
	if err != nil {
		return nil, err
	}

	return &orderScanner{
		rows:    rows,
		indexes: indexes,
		mark:    -1,
	}, nil
}

func (s *orderScanner) Scan() (shared.Row, error) {
	if s.next < len(s.indexes) {
		row := s.rows.Row(s.indexes[s.next])
		s.next++
		return row, nil
	}

	return shared.Row{}, scan.ErrNoRows
}

func (s *orderScanner) Mark() {
	s.mark = s.next - 1
}

func (s *orderScanner) Restore() {
	if s.mark == -1 {
		panic("no mark to restore")
	}

	s.next = s.mark
}

func findIndexIterationOrder(ctx queries.Context, order expressions.OrderExpression, rows shared.Rows) ([]int, error) {
	var expressions []expressions.ExpressionWithDirection
	if order != nil {
		expressions = order.Expressions()
	}

	indexValues, err := makeIndexValues(ctx, expressions, rows)
	if err != nil {
		return nil, err
	}

	incomparable := false
	sort.SliceStable(indexValues, func(i, j int) bool {
		for k, value := range indexValues[i].values {
			reverse := expressions[k].Reverse

			switch shared.CompareValues(value, indexValues[j].values[k]) {
			case shared.OrderTypeIncomparable:
				incomparable = true
				return false
			case shared.OrderTypeBefore:
				return !reverse
			case shared.OrderTypeAfter:
				return reverse
			}
		}

		return false
	})
	if incomparable {
		return nil, fmt.Errorf("incomparable types")
	}

	indexes := make([]int, 0, len(indexValues))
	for _, value := range indexValues {
		indexes = append(indexes, value.index)
	}

	return indexes, nil
}

type indexValue struct {
	index  int
	values []any
}

func makeIndexValues(ctx queries.Context, expressions []expressions.ExpressionWithDirection, rows shared.Rows) ([]indexValue, error) {
	indexValues := make([]indexValue, 0, len(rows.Values))
	for i := range rows.Values {
		values := make([]any, 0, len(expressions))
		for _, expression := range expressions {
			value, err := queries.Evaluate(ctx, expression.Expression, rows.Row(i))
			if err != nil {
				return nil, err
			}

			values = append(values, value)
		}

		indexValues = append(indexValues, indexValue{index: i, values: values})
	}

	return indexValues, nil
}
