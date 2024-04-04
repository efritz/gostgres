package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type Node interface {
	Name() string
	Fields() []shared.Field
	Serialize(w io.Writer, indentationLevel int)
	Optimize()
	AddFilter(filter expressions.Expression)
	AddOrder(order OrderExpression)
	Filter() expressions.Expression
	Ordering() OrderExpression

	// TODO: rough implementation
	// https://sourcegraph.com/github.com/postgres/postgres@06286709ee0637ec7376329a5aa026b7682dcfe2/-/blob/src/backend/executor/execAmi.c?L439:59-439:79
	SupportsMarkRestore() bool
	Scanner() (Scanner, error)
}

type VisitorFunc func(row shared.Row) (bool, error)

type Scanner interface {
	Scan() (shared.Row, error)
}

type MarkRestorer interface {
	Mark()
	Restore()
}

var ErrNoRows = fmt.Errorf("no rows")

type ScannerFunc func() (shared.Row, error)

func (f ScannerFunc) Scan() (shared.Row, error) {
	return f()
}

func Empty() Scanner {
	return ScannerFunc(func() (shared.Row, error) {
		return shared.Row{}, ErrNoRows
	})
}

func ScanRows(node Node) (shared.Rows, error) {
	scanner, err := node.Scanner()
	if err != nil {
		return shared.Rows{}, err
	}

	rows, err := shared.NewRows(node.Fields())
	if err != nil {
		return shared.Rows{}, err
	}

	if err := VisitRows(scanner, func(row shared.Row) (bool, error) {
		rows, err = rows.AddValues(row.Values)
		return true, err
	}); err != nil {
		return shared.Rows{}, err
	}

	return rows, nil
}

func VisitRows(scanner Scanner, visitor VisitorFunc) error {
	for {
		row, err := scanner.Scan()
		if err != nil {
			if err == ErrNoRows {
				break
			}

			return err
		}

		if ok, err := visitor(row); err != nil {
			return err
		} else if !ok {
			break
		}
	}

	return nil
}
