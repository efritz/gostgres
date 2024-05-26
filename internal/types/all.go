package types

import (
	"fmt"

	"github.com/efritz/gostgres/internal/catalog"
	"github.com/efritz/gostgres/internal/shared"
)

type Table interface {
	Name() string
	Indexes() []BaseIndex
	Fields() []TableField
	Size() int
	TIDs() []int64
	Row(tid int64) (shared.Row, bool)
	SetPrimaryKey(index BaseIndex) error
	AddIndex(index BaseIndex) error
	AddConstraint(ctx Context, constraint Constraint) error
	Insert(ctx Context, row shared.Row) (_ shared.Row, err error)
	Delete(row shared.Row) (shared.Row, bool, error)
}

type BaseIndex interface {
	Name() string
	Unwrap() BaseIndex
	UniqueOn() []shared.Field
	Filter() Expression
	Insert(row shared.Row) error
	Delete(row shared.Row) error
}

type Constraint interface {
	Name() string
	Check(ctx Context, row shared.Row) error
}

type Function interface {
	Invoke(ctx Context, args []any) (any, error)
}

type Aggregate interface {
	Step(state any, args []any) (any, error)
	Done(state any) (any, error)
}

type Sequence interface {
	Name() string
	Next() (int64, error)
	Set(value int64) error
	Value() int64
}

type Expression interface {
	fmt.Stringer

	Equal(other Expression) bool
	Fold() Expression
	Map(f func(Expression) Expression) Expression
	ValueFrom(cts Context, row shared.Row) (any, error)
}

type Context struct {
	tables     *catalog.Catalog[Table]
	sequences  *catalog.Catalog[Sequence]
	functions  *catalog.Catalog[Function]
	aggregates *catalog.Catalog[Aggregate]
	outerRow   shared.Row
}

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

var EmptyContext = NewContext(
	catalog.NewCatalog[Table](),
	catalog.NewCatalog[Sequence](),
	catalog.NewCatalog[Function](),
	catalog.NewCatalog[Aggregate](),
)
