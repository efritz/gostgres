package relations

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type filterRelation struct {
	Relation
	filter expressions.BoolExpression
}

var _ Relation = &filterRelation{}

func NewFilter(table Relation, filter expressions.BoolExpression) Relation {
	return &filterRelation{
		Relation: table,
		filter:   filter,
	}
}

func (r *filterRelation) Scan(visitor VisitorFunc) error {
	return r.Relation.Scan(func(row shared.Row) (bool, error) {
		if ok, err := r.filter.ValueFrom(row); err != nil {
			return false, err
		} else if !ok {
			return true, nil
		}

		return visitor(row)
	})
}
