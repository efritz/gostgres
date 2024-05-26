package indexes

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/types"
)

type Index[O ScanOptions] interface {
	types.BaseIndex

	Description(opts O) string
	Condition(opts O) types.Expression
	Ordering(opts O) expressions.OrderExpression
	Scanner(ctx types.Context, opts O) (tidScanner, error)
}

type ScanOptions any
