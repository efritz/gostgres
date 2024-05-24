package eval

import (
	"github.com/efritz/gostgres/internal/catalog/aggregates"
	"github.com/efritz/gostgres/internal/catalog/functions"
	"github.com/efritz/gostgres/internal/catalog/sequence"
	"github.com/efritz/gostgres/internal/catalog/table"
	"github.com/efritz/gostgres/internal/shared"
)

type Context struct {
	tables     *Catalog[*table.Table]
	sequences  *Catalog[*sequence.Sequence]
	functions  *Catalog[functions.Function]
	aggregates *Catalog[aggregates.Aggregate]
	outerRow   shared.Row
}

func NewContext(
	tables *Catalog[*table.Table],
	sequences *Catalog[*sequence.Sequence],
	functions *Catalog[functions.Function],
	aggregates *Catalog[aggregates.Aggregate],
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
	return ctx.tables.Get(name)
}

func (ctx Context) GetFunction(name string) (functions.Function, bool) {
	return ctx.functions.Get(name)
}

func (ctx Context) GetSequence(name string) (*sequence.Sequence, bool) {
	return ctx.sequences.Get(name)
}

func (ctx Context) GetAggregate(name string) (aggregates.Aggregate, bool) {
	return ctx.aggregates.Get(name)
}

func (ctx Context) CreateTable(name string, fields []table.TableField) error {
	ctx.tables.Set(name, table.NewTable(name, fields))
	return nil
}

func (ctx Context) CreateAndGetSequence(name string, typ shared.Type) (*sequence.Sequence, error) {
	sequence := sequence.NewSequence(name, typ)
	ctx.sequences.Set(name, sequence)
	return sequence, nil
}
