package types

import "github.com/efritz/gostgres/internal/shared"

type BaseIndex interface {
	Name() string
	Unwrap() BaseIndex
	UniqueOn() []shared.Field
	Filter() Expression
	Insert(row shared.Row) error
	Delete(row shared.Row) error
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
