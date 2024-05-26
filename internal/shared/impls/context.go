package impls

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/efritz/gostgres/internal/catalog"
	"github.com/efritz/gostgres/internal/shared/rows"
)

type Context struct {
	tables     *catalog.Catalog[Table]
	sequences  *catalog.Catalog[Sequence]
	functions  *catalog.Catalog[Function]
	aggregates *catalog.Catalog[Aggregate]
	outerRow   rows.Row
	debug      bool
}

var EmptyContext = NewContext(
	catalog.NewCatalog[Table](),
	catalog.NewCatalog[Sequence](),
	catalog.NewCatalog[Function](),
	catalog.NewCatalog[Aggregate](),
)

func NewContext(
	tables *catalog.Catalog[Table],
	sequences *catalog.Catalog[Sequence],
	functions *catalog.Catalog[Function],
	aggregates *catalog.Catalog[Aggregate],
) Context {
	return Context{
		tables:     tables,
		sequences:  sequences,
		functions:  functions,
		aggregates: aggregates,
	}
}

func (c Context) OuterRow() rows.Row {
	return c.outerRow
}

func (c Context) WithOuterRow(row rows.Row) Context {
	return Context{
		tables:     c.tables,
		sequences:  c.sequences,
		functions:  c.functions,
		aggregates: c.aggregates,
		outerRow:   row,
	}
}

func (ctx Context) GetTable(name string) (Table, bool) {
	return ctx.tables.Get(name)
}

func (ctx Context) GetFunction(name string) (Function, bool) {
	return ctx.functions.Get(name)
}

func (ctx Context) GetSequence(name string) (Sequence, bool) {
	return ctx.sequences.Get(name)
}

func (ctx Context) GetAggregate(name string) (Aggregate, bool) {
	return ctx.aggregates.Get(name)
}

func (ctx Context) SetTable(name string, table Table) {
	ctx.tables.Set(name, table)
}

func (ctx Context) SetSequence(name string, sequence Sequence) {
	ctx.sequences.Set(name, sequence)
}

func (c Context) WithDebug() Context {
	c.debug = true
	return c
}

func (c Context) Log(format string, args ...interface{}) {
	if !c.debug {
		return
	}

	_, file, line, _ := runtime.Caller(1)
	parts := strings.Split(file, "/gostgres/internal/execution/queries/")
	caller := fmt.Sprintf("[%s:%d]", parts[1], line)

	fmt.Printf("%% [%s] ", caller)
	fmt.Printf(format, args...)
	fmt.Printf("\n")
}
