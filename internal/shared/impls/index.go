package impls

import (
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/rows"
)

type BaseIndex interface {
	Name() string
	Unwrap() BaseIndex
	UniqueOn() []fields.Field
	Filter() Expression
	Insert(row rows.Row) error
	Delete(row rows.Row) error
}

type ScanOptions any

type Index[O ScanOptions] interface {
	BaseIndex

	Description(opts O) string
	Condition(opts O) Expression
	Ordering(opts O) OrderExpression
	Scanner(ctx Context, opts O) (TIDScanner, error)
}

type TIDScanner interface {
	Scan() (int64, error)
}
