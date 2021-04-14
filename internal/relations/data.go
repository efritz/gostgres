package relations

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type dataRelation struct {
	name string
	rows shared.Rows

	// TODO - sub selection
	filter expressions.BoolExpression
	order  expressions.IntExpression
}

var _ Relation = &dataRelation{}

func NewData(name string, rows shared.Rows) Relation {
	return &dataRelation{
		name: name,
		rows: rows,
	}
}

func NewDataWithFilters(name string, rows shared.Rows, filter expressions.BoolExpression, order expressions.IntExpression) Relation {
	return &dataRelation{
		name:   name,
		rows:   rows,
		filter: filter,
		order:  order,
	}
}

func (r *dataRelation) Name() string           { return r.name }
func (r *dataRelation) Fields() []shared.Field { return copyFields(r.rows.Fields) }

func (r *dataRelation) Scan(visitor VisitorFunc) error {
	indexes, err := findIndexIterationOrder(r.order, r.rows)
	if err != nil {
		return err
	}

	for _, i := range indexes {
		row := r.rows.Row(i)

		if r.filter != nil {
			if ok, err := r.filter.ValueFrom(row); err != nil {
				return err
			} else if !ok {
				continue
			}
		}

		if ok, err := visitor(row); err != nil {
			return err
		} else if !ok {
			break
		}
	}

	return nil
}
