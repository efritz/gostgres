package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type indexAccessStrategy[O ScanOptions] struct {
	table *Table
	index Index[O]
	opts  O
}

var _ accessStrategy = &indexAccessStrategy[ScanOptions]{}

func NewIndexAccessStrategy[O ScanOptions](table *Table, index Index[O], opts O) accessStrategy {
	return &indexAccessStrategy[O]{
		table: table,
		index: index,
		opts:  opts,
	}
}

func (s *indexAccessStrategy[ScanOptions]) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%sindex scan of %s via %s\n", indent(indentationLevel), s.table.name, s.index.Name()))

	if filter := s.Filter(); filter != nil {
		io.WriteString(w, fmt.Sprintf("%sindex cond: %s\n", indent(indentationLevel+1), filter))
	}
}

func (s *indexAccessStrategy[ScanOptions]) Filter() expressions.Expression {
	return s.index.Condition(s.opts)
}

func (s *indexAccessStrategy[ScanOptions]) Ordering() OrderExpression {
	return s.index.Ordering()
}

func (s *indexAccessStrategy[ScanOptions]) Scanner(ctx ScanContext) (Scanner, error) {
	tidScanner, err := s.index.Scanner(ctx, s.opts)
	if err != nil {
		return nil, err
	}

	return ScannerFunc(func() (shared.Row, error) {
		tid, err := tidScanner.Scan()
		if err != nil {
			return shared.Row{}, err
		}

		row := s.table.rows[tid]
		return row, nil
	}), nil
}
