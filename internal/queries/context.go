package queries

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/functions"
	"github.com/efritz/gostgres/internal/sequence"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/table"
)

type Context struct {
	Tables    *table.Tablespace
	Sequences *sequence.Sequencespace
	Functions *functions.Functionspace
	OuterRow  shared.Row
}

func NewContext(tables *table.Tablespace, Sequences *sequence.Sequencespace, functions *functions.Functionspace) Context {
	return Context{
		Tables:    tables,
		Sequences: Sequences,
		Functions: functions,
	}
}

func (c Context) WithOuterRow(row shared.Row) Context {
	return Context{
		Tables:   c.Tables,
		OuterRow: row,
	}
}

func (ctx Context) Evaluate(expr expressions.Expression, row shared.Row) (any, error) {
	return expr.ValueFrom(ctx, shared.CombineRows(row, ctx.OuterRow))
}

func (ctx Context) GetFunction(name string) (functions.Function, bool) {
	return ctx.Functions.GetFunction(name)
}

func (ctx Context) GetSequence(name string) (*sequence.Sequence, bool) {
	return ctx.Sequences.GetSequence(name)
}
