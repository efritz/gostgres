package impls

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/efritz/gostgres/internal/catalog"
	"github.com/efritz/gostgres/internal/shared/rows"
)

type Context struct {
	Tables     *catalog.Catalog[Table]
	Sequences  *catalog.Catalog[Sequence]
	Functions  *catalog.Catalog[Function]
	Aggregates *catalog.Catalog[Aggregate]
	debug      bool
	outerRow   rows.Row
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
		Tables:     tables,
		Sequences:  sequences,
		Functions:  functions,
		Aggregates: aggregates,
	}
}

func (c Context) OuterRow() rows.Row                { return c.outerRow }
func (c Context) WithDebug() Context                { c.debug = true; return c }
func (c Context) WithOuterRow(row rows.Row) Context { c.outerRow = row; return c }

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
