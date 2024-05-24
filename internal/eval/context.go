package eval

import (
	"github.com/efritz/gostgres/internal/aggregates"
	"github.com/efritz/gostgres/internal/functions"
	"github.com/efritz/gostgres/internal/sequence"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/table"
)

type Context struct {
	tables     *table.Tablespace
	sequences  *sequence.Sequencespace
	functions  *functions.Functionspace
	aggregates *aggregates.Aggregatespace
	outerRow   shared.Row
}

func NewContext(
	tables *table.Tablespace,
	sequences *sequence.Sequencespace,
	functions *functions.Functionspace,
	aggregates *aggregates.Aggregatespace,
) Context {
	return Context{
		tables:     tables,
		sequences:  sequences,
		functions:  functions,
		aggregates: aggregates,
	}
}

func (c Context) OuterRow() shared.Row {
	return c.outerRow
}

func (c Context) WithOuterRow(row shared.Row) Context {
	return Context{
		tables:     c.tables,
		sequences:  c.sequences,
		functions:  c.functions,
		aggregates: c.aggregates,
		outerRow:   row,
	}
}

func (ctx Context) GetTable(name string) (*table.Table, bool) {
	return ctx.tables.GetTable(name)
}

func (ctx Context) CreateTable(name string, fields []table.TableField) error {
	return ctx.tables.CreateTable(name, fields)
}

func (ctx Context) GetFunction(name string) (functions.Function, bool) {
	return ctx.functions.GetFunction(name)
}

func (ctx Context) GetSequence(name string) (*sequence.Sequence, bool) {
	return ctx.sequences.GetSequence(name)
}

func (ctx Context) CreateAndGetSequence(name string, typ shared.Type) (*sequence.Sequence, error) {
	return ctx.sequences.CreateAndGetSequence(name, typ)
}

func (ctx Context) GetAggregate(name string) (aggregates.Aggregate, bool) {
	return ctx.aggregates.GetAggregate(name)
}
