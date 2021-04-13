package relations

import (
	"github.com/efritz/gostgres/internal/shared"
)

type Relation interface {
	Name() string
	Fields() []shared.Field
	Scan(scanContext ScanContext, visitor VisitorFunc) error
}

type VisitorFunc func(scanContext ScanContext, values []interface{}) (bool, error)

type ScanContext struct {
	// TODO - selection, filtering, ordering
}

func ScanRows(scanContext ScanContext, relation Relation) (shared.Rows, error) {
	rows := shared.Rows{
		Fields: relation.Fields(),
	}

	if err := relation.Scan(scanContext, func(scanContext ScanContext, values []interface{}) (bool, error) {
		rows.Values = append(rows.Values, values)
		return true, nil
	}); err != nil {
		return shared.Rows{}, err
	}

	return rows, nil
}
