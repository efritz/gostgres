package relations

import (
	"fmt"
	"sort"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type indexValue struct {
	index int
	value interface{}
}

func findIndexIterationOrder(order expressions.Expression, rows shared.Rows) ([]int, error) {
	indexValues := make([]indexValue, 0, len(rows.Values))
	for i := range rows.Values {
		value, err := indexValueFrom(order, i, rows.Row(i))
		if err != nil {
			return nil, err
		}

		indexValues = append(indexValues, value)
	}

	incomparable := false
	sort.Slice(indexValues, func(i, j int) bool {
		relation := shared.CompareValues(indexValues[i].value, indexValues[j].value)
		if relation == shared.OrderTypeIncomparable {
			incomparable = true
		}
		return relation == shared.OrderTypeBefore
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

func indexValueFrom(expression expressions.Expression, index int, row shared.Row) (indexValue, error) {
	if expression == nil {
		return indexValue{index: index, value: index}, nil
	}

	value, err := expression.ValueFrom(row)
	if err != nil {
		return indexValue{}, err
	}

	return indexValue{index: index, value: value}, nil
}
