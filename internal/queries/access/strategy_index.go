package access

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/indexes"
	"github.com/efritz/gostgres/internal/queries/filter"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
)

type indexAccessStrategy[O indexes.ScanOptions] struct {
	table indexes.TableIndexer
	index indexes.Index[O]
	opts  O
}

var _ accessStrategy = &indexAccessStrategy[indexes.ScanOptions]{}

func NewIndexAccessStrategy[O indexes.ScanOptions](table indexes.TableIndexer, index indexes.Index[O], opts O) accessStrategy {
	return &indexAccessStrategy[O]{
		table: table,
		index: index,
		opts:  opts,
	}
}

func (s *indexAccessStrategy[ScanOptions]) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%s%s\n", serialization.Indent(indentationLevel), s.index.Description(s.opts)))

	if filter := s.Filter(); filter != nil {
		io.WriteString(w, fmt.Sprintf("%sindex cond: %s\n", serialization.Indent(indentationLevel+1), filter))
	}
}

func (s *indexAccessStrategy[ScanOptions]) Filter() expressions.Expression {
	filterExpression := s.index.Filter()
	condition := s.index.Condition(s.opts)

	if filterExpression == nil {
		return condition
	}
	if condition == nil {
		return filterExpression
	}

	return filter.UnionFilters(append(filterExpression.Conjunctions(), condition.Conjunctions()...)...)
}

func (s *indexAccessStrategy[ScanOptions]) Ordering() expressions.OrderExpression {
	return s.index.Ordering(s.opts)
}

func (s *indexAccessStrategy[ScanOptions]) Scanner(ctx scan.ScanContext) (scan.Scanner, error) {
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
