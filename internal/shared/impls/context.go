package impls

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/efritz/gostgres/internal/shared/rows"
)

type ExpressionResolutionContext struct {
	Catalog CatalogSet
}

func NewExpressionResolutionContext(catalog CatalogSet) ExpressionResolutionContext {
	return ExpressionResolutionContext{
		Catalog: catalog,
	}
}

//
//

type NodeResolutionContext struct {
	Catalog CatalogSet
}

func NewNodeResolutionContext(catalog CatalogSet) *NodeResolutionContext {
	return &NodeResolutionContext{
		Catalog: catalog,
	}
}

func (ctx *NodeResolutionContext) ExpressionResolutionContext() ExpressionResolutionContext {
	return NewExpressionResolutionContext(ctx.Catalog)
}

//
//

type ExecutionContext struct {
	Catalog  CatalogSet
	debug    bool
	outerRow rows.Row
}

var EmptyExecutionContext = NewExecutionContext(NewCatalogEmptySet())

func NewExecutionContext(catalog CatalogSet) ExecutionContext {
	return ExecutionContext{
		Catalog: catalog,
	}
}

func (c ExecutionContext) OuterRow() rows.Row                         { return c.outerRow }
func (c ExecutionContext) WithDebug() ExecutionContext                { c.debug = true; return c }
func (c ExecutionContext) WithOuterRow(row rows.Row) ExecutionContext { c.outerRow = row; return c }

func (c ExecutionContext) Log(format string, args ...interface{}) {
	if !c.debug {
		return
	}

	_, file, line, _ := runtime.Caller(1)
	parts := strings.Split(file, "/gostgres/internal/")

	fmt.Printf("%% [%s:%d] ", parts[1], line)
	fmt.Printf(format, args...)
	fmt.Printf("\n")
}
