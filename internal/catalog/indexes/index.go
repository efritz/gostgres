package indexes

import (
	"github.com/efritz/gostgres/internal/catalog/table"
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
)

type Index[O ScanOptions] interface {
	table.Index

	Description(opts O) string
	Condition(opts O) expressions.Expression
	Ordering(opts O) expressions.OrderExpression
	Scanner(ctx queries.Context, opts O) (tidScanner, error)
}

type ScanOptions any
