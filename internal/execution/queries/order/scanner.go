package order

import (
	"fmt"
	"sort"

	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/ordering"
	"github.com/efritz/gostgres/internal/shared/rows"
)

type orderScanner struct {
	ctx     impls.Context
	rows    rows.Rows
	indexes []int
	next    int
	mark    int
}

func NewOrderScanner(ctx impls.Context, scanner scan.Scanner, fields []fields.Field, order impls.OrderExpression) (scan.Scanner, error) {
	ctx.Log("Building Order scanner")

	rows, err := rows.NewRows(fields)
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
		ctx:     ctx,
		rows:    rows,
		indexes: indexes,
		mark:    -1,
	}, nil
}

func (s *orderScanner) Scan() (rows.Row, error) {
	s.ctx.Log("Scanning Order")

	if s.next < len(s.indexes) {
		row := s.rows.Row(s.indexes[s.next])
		s.next++
		return row, nil
	}

	return rows.Row{}, scan.ErrNoRows
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

func findIndexIterationOrder(ctx impls.Context, order impls.OrderExpression, rows rows.Rows) ([]int, error) {
	var expressions []impls.ExpressionWithDirection
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

			switch ordering.CompareValues(value, indexValues[j].values[k]) {
			case ordering.OrderTypeIncomparable:
				incomparable = true
				return false
			case ordering.OrderTypeBefore:
				return !reverse
			case ordering.OrderTypeAfter:
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

func makeIndexValues(ctx impls.Context, expressions []impls.ExpressionWithDirection, rows rows.Rows) ([]indexValue, error) {
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
