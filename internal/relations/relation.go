package relations

import (
	"bytes"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type Relation interface {
	Name() string
	Fields() []shared.Field
	Serialize(buf *bytes.Buffer, indentationLevel int)
	Optimize()
	SinkFilter(filter expressions.Expression) bool
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
