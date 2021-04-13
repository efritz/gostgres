package relations

import (
	"github.com/efritz/gostgres/internal/shared"
)

type Relation interface {
	Name() string
	Fields() []shared.Field
	Scan(visitor VisitorFunc) error
}

type VisitorFunc func(row shared.Row) (bool, error)

func ScanRows(relation Relation) (shared.Rows, error) {
	rows := shared.NewRows(relation.Fields())

	if err := relation.Scan(func(row shared.Row) (bool, error) {
		rows = rows.AddValues(row.Values)
		return true, nil
	}); err != nil {
		return rows, err
	}

	return rows, nil
}
