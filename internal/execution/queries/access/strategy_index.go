package access

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type indexAccessStrategy[O impls.ScanOptions] struct {
	table impls.Table
	index impls.Index[O]
	opts  O
}

var _ accessStrategy = &indexAccessStrategy[impls.ScanOptions]{}

func NewIndexAccessStrategy[O impls.ScanOptions](table impls.Table, index impls.Index[O], opts O) accessStrategy {
	return &indexAccessStrategy[O]{
		table: table,
		index: index,
		opts:  opts,
	}
}

func (s *indexAccessStrategy[ScanOptions]) Serialize(w serialization.IndentWriter) {
	w.WritefLine(s.index.Description(s.opts))

	if filter := s.Filter(); filter != nil {
		w.Indent().WritefLine("index cond: %s", filter)
	}
}

func (s *indexAccessStrategy[ScanOptions]) Filter() impls.Expression {
	filterExpression := s.index.Filter()
	condition := s.index.Condition(s.opts)

	if filterExpression == nil {
		return condition
	}
	if condition == nil {
		return filterExpression
	}

	return expressions.UnionFilters(append(expressions.Conjunctions(filterExpression), expressions.Conjunctions(condition)...)...)
}

func (s *indexAccessStrategy[ScanOptions]) Ordering() impls.OrderExpression {
	return s.index.Ordering(s.opts)
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
