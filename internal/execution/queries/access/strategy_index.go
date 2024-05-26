package access

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

type indexAccessStrategy[O types.ScanOptions] struct {
	table types.Table
	index types.Index[O]
	opts  O
}

var _ accessStrategy = &indexAccessStrategy[types.ScanOptions]{}

func NewIndexAccessStrategy[O types.ScanOptions](table types.Table, index types.Index[O], opts O) accessStrategy {
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

func (s *indexAccessStrategy[ScanOptions]) Filter() types.Expression {
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

func (s *indexAccessStrategy[ScanOptions]) Ordering() types.OrderExpression {
	return s.index.Ordering(s.opts)
}

func (s *indexAccessStrategy[ScanOptions]) Scanner(ctx types.Context) (scan.Scanner, error) {
	tidScanner, err := s.index.Scanner(ctx, s.opts)
	if err != nil {
		return nil, err
	}

	return scan.ScannerFunc(func() (shared.Row, error) {
		tid, err := tidScanner.Scan()
		if err != nil {
			return shared.Row{}, err
		}

		row, ok := s.table.Row(tid)
		if !ok {
			return shared.Row{}, fmt.Errorf("row not found for tid %d", tid)
		}

		return row, nil
	}), nil
}
