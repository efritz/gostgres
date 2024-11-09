package impls

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/efritz/gostgres/internal/shared/rows"
)

type Context struct {
	Catalog  CatalogSet
	debug    bool
	outerRow rows.Row
}

var EmptyContext = NewContext(NewCatalogEmptySet())

func NewContext(catalog CatalogSet) Context {
	return Context{
		Catalog: catalog,
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
	parts := strings.Split(file, "/gostgres/internal/")

	fmt.Printf("%% [%s:%d] ", parts[1], line)
	fmt.Printf(format, args...)
	fmt.Printf("\n")
}
