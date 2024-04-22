package queries

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/table"
)

type Context struct {
	Tables   *table.Tablespace
	OuterRow shared.Row
}

func NewContext(tables *table.Tablespace) Context {
	return Context{Tables: tables}
}

func (c Context) WithOuterRow(row shared.Row) Context {
	return Context{
		Tables:   c.Tables,
		OuterRow: row,
	}
}

func (ctx Context) Evaluate(expr expressions.Expression, row shared.Row) (any, error) {
	return expr.ValueFrom(shared.CombineRows(row, ctx.OuterRow))
}
