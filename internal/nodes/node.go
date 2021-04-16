package nodes

import (
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type Node interface {
	Name() string
	Fields() []shared.Field
	Serialize(w io.Writer, indentationLevel int)
	Optimize()
	PushDownFilter(filter expressions.Expression) bool
	Scan(visitor VisitorFunc) error
}

type VisitorFunc func(row shared.Row) (bool, error)

func ScanRows(node Node) (shared.Rows, error) {
	rows := shared.NewRows(node.Fields())

	if err := node.Scan(func(row shared.Row) (_ bool, err error) {
		rows, err = rows.AddValues(row.Values)
		return true, err
	}); err != nil {
		return rows, err
	}

	return rows, nil
}
