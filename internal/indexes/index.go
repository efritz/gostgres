package indexes

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
)

type BaseIndex interface {
	Unwrap() BaseIndex
	Filter() expressions.Expression
	Insert(row shared.Row) error
	Delete(row shared.Row) error
}

type ScanOptions any

type IndexScanner[O ScanOptions] interface {
	Scanner(ctx scan.ScanContext, opts O) (tidScanner, error)
}

type Index[O ScanOptions] interface {
	BaseIndex
	IndexScanner[O]

	Description(opts O) string
	Condition(opts O) expressions.Expression
	Ordering(opts O) expressions.OrderExpression
}

type TableIndexer interface {
	Row(tid int) (shared.Row, bool)
}
