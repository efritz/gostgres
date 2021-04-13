package relations

import (
	"sort"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type orderRelation struct {
	Relation
	expression expressions.Expression
}

var _ Relation = &orderRelation{}

func NewOrder(relation Relation, expression expressions.Expression) Relation {
	return &orderRelation{
		Relation:   relation,
		expression: expression,
	}
}

func (r *orderRelation) Scan(scanContext ScanContext, visitor VisitorFunc) error {
	rows, err := ScanRows(scanContext, r.Relation)
	if err != nil {
		return err
	}

	type sortableValue struct {
		index int
		value int
	}

	sortableValues := make([]sortableValue, 0, len(rows.Values))
	for i, rowValues := range rows.Values {
		value, err := r.expression.ValueFrom(shared.Row{Fields: rows.Fields, Values: rowValues})
		if err != nil {
			return err
		}

		intValue, ok := value.(int)
		if !ok {
			panic("non-int ordering is currently unsupported")
		}

		sortableValues = append(sortableValues, sortableValue{i, intValue})
	}

	sort.Slice(sortableValues, func(i, j int) bool {
		return sortableValues[i].value < sortableValues[j].value
	})

	for _, value := range sortableValues {
		if ok, err := visitor(scanContext, rows.Values[value.index]); err != nil {
			return err
		} else if !ok {
			break
		}
	}

	return nil
}
