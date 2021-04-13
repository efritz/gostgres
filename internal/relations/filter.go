package relations

import (
	"github.com/efritz/gostgres/internal/filters"
	"github.com/efritz/gostgres/internal/shared"
)

type filterRelation struct {
	Relation
	filter filters.Filter
}

var _ Relation = &filterRelation{}

func NewFilter(table Relation, filter filters.Filter) Relation {
	return &filterRelation{
		Relation: table,
		filter:   filter,
	}
}

func (r *filterRelation) Scan(scanContext ScanContext, visitor VisitorFunc) error {
	fields := r.Fields()

	return r.Relation.Scan(scanContext, func(scanContext ScanContext, values []interface{}) (bool, error) {
		if ok, err := r.filter.Test(shared.Row{Fields: fields, Values: values}); err != nil {
			return false, err
		} else if !ok {
			return true, nil
		}

		return visitor(scanContext, values)
	})
}
