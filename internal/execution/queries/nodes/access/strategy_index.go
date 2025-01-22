package access

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type indexAccessStrategy[O impls.ScanOptions] struct {
	table     impls.Table
	index     impls.Index[O]
	opts      O
	condition impls.Expression
}

func NewIndexAccessStrategy[O impls.ScanOptions](table impls.Table, index impls.Index[O], opts O, condition impls.Expression) nodes.AccessStrategy {
	return &indexAccessStrategy[O]{
		table:     table,
		index:     index,
		opts:      opts,
		condition: condition,
	}
}

func (s *indexAccessStrategy[ScanOptions]) Serialize(w serialization.IndentWriter) {
	w.WritefLine(s.index.Description(s.opts))

	if s.condition != nil {
		w.Indent().WritefLine("index cond: %s", s.condition)
	}
}

func (s *indexAccessStrategy[ScanOptions]) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Index Access scanner Strategy")

	tidScanner, err := s.index.Scanner(ctx, s.opts)
	if err != nil {
		return nil, err
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Index Access Strategy")

		tid, err := tidScanner.Scan()
		if err != nil {
			return rows.Row{}, err
		}

		row, ok := s.table.Row(tid)
		if !ok {
			return rows.Row{}, fmt.Errorf("row not found for tid %d", tid)
		}

		return row, nil
	}), nil
}
