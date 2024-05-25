package indexes

import (
	"github.com/efritz/gostgres/internal/catalog/table"
	"github.com/efritz/gostgres/internal/execution"
	"github.com/efritz/gostgres/internal/execution/expressions"
)

type BaseIndex = table.Index

type Index[O ScanOptions] interface {
	BaseIndex

	Description(opts O) string
	Condition(opts O) expressions.Expression
	Ordering(opts O) expressions.OrderExpression
	Scanner(ctx execution.Context, opts O) (tidScanner, error)
}

type ScanOptions any
