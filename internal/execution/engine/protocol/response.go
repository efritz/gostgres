package protocol

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/rows"
)

type ResponseWriter interface {
	SendRow(row rows.Row)
	Done()
	Error(err error)
}

type rowCollector struct {
	done bool
	rows rows.Rows
	err  error
}

var _ ResponseWriter = &rowCollector{}

func NewRowCollector() *rowCollector {
	return &rowCollector{}
}

func (w *rowCollector) Rows() (rows.Rows, error) {
	if !w.done {
		return rows.Rows{}, fmt.Errorf("not done")
	}

	return w.rows, w.err
}

func (w *rowCollector) SendRow(row rows.Row) {
	if len(w.rows.Fields) == 0 {
		w.rows.Fields = row.Fields
		w.rows.Values = [][]any{row.Values}
		return
	}

	if rows, err := w.rows.AddValues(row.Values); err != nil {
		w.Error(err)
	} else {
		w.rows = rows
	}
}

func (w *rowCollector) Done() {
	w.done = true
}

func (w *rowCollector) Error(err error) {
	w.done = true
	w.err = err
}
