package protocol

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type ResponseWriter interface {
	SendRow(row shared.Row)
	Done()
	Error(err error)
}

type rowCollector struct {
	done bool
	rows shared.Rows
	err  error
}

var _ ResponseWriter = &rowCollector{}

func NewRowCollector() *rowCollector {
	return &rowCollector{}
}

func (w *rowCollector) Rows() (shared.Rows, error) {
	if !w.done {
		return shared.Rows{}, fmt.Errorf("not done")
	}

	return w.rows, w.err
}

func (w *rowCollector) SendRow(row shared.Row) {
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
