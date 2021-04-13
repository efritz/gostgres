package relations

import (
	"sort"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type orderRelation struct {
	Relation
	order expressions.IntExpression
}

var _ Relation = &orderRelation{}

func NewOrder(relation Relation, expression expressions.IntExpression) Relation {
	return &orderRelation{
		Relation: relation,
		order:    expression,
	}
}

func (r *orderRelation) Scan(visitor VisitorFunc) error {
	rows, err := ScanRows(r.Relation)
	if err != nil {
		return err
	}

	indexes, err := findIndexIterationOrder(r.order, rows)
	if err != nil {
		return err
	}

	for _, i := range indexes {
		if ok, err := visitor(rows.Row(i)); err != nil {
			return err
		} else if !ok {
			break
		}
	}

	return nil
}

type indexValue struct {
	index int
	value int
}

func findIndexIterationOrder(order expressions.IntExpression, rows shared.Rows) ([]int, error) {
	indexValues := make([]indexValue, 0, len(rows.Values))
	for i := range rows.Values {
		value, err := indexValueFrom(order, i, rows.Row(i))
		if err != nil {
			return nil, err
		}

		indexValues = append(indexValues, value)
	}

	sort.Slice(indexValues, func(i, j int) bool {
		return indexValues[i].value < indexValues[j].value
	})

	indexes := make([]int, 0, len(indexValues))
	for _, value := range indexValues {
		indexes = append(indexes, value.index)
	}

	return indexes, nil
}

func indexValueFrom(expression expressions.IntExpression, index int, row shared.Row) (indexValue, error) {
	if expression == nil {
		return indexValue{index: index, value: index}, nil
	}

	value, err := expression.ValueFrom(row)
	if err != nil {
		return indexValue{}, err
	}

	return indexValue{index: index, value: value}, nil
}
