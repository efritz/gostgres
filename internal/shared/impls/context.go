package impls

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/rows"
)

type Cataloger interface {
	Catalog() CatalogSet
}

//
//

type ExpressionResolutionContext struct {
	catalog                 CatalogSet
	allowAggregateFunctions bool
}

func NewExpressionResolutionContext(catalog CatalogSet, allowAggregateFunctions bool) ExpressionResolutionContext {
	return ExpressionResolutionContext{
		catalog:                 catalog,
		allowAggregateFunctions: allowAggregateFunctions,
	}
}

func (ctx ExpressionResolutionContext) Catalog() CatalogSet {
	return ctx.catalog
}

func (ctx ExpressionResolutionContext) AllowAggregateFunctions() bool {
	return ctx.allowAggregateFunctions
}

//
//

type NodeResolutionContext struct {
	catalog CatalogSet
	Scopes  []Scope
}

type Scope struct {
	fields []fields.Field
}

func NewNodeResolutionContext(catalog CatalogSet) *NodeResolutionContext {
	return &NodeResolutionContext{
		catalog: catalog,
	}
}

func (ctx *NodeResolutionContext) Catalog() CatalogSet {
	return ctx.catalog
}

func (ctx *NodeResolutionContext) ExpressionResolutionContext(allowAggregateFunctions bool) ExpressionResolutionContext {
	return NewExpressionResolutionContext(ctx.catalog, allowAggregateFunctions)
}

func (ctx *NodeResolutionContext) WithScope(f func() error) error {
	ctx.PushScope()
	defer ctx.PopScope()

	return f()
}

func (ctx *NodeResolutionContext) PushScope() {
	ctx.Scopes = append(ctx.Scopes, Scope{})
}

func (ctx *NodeResolutionContext) PopScope() {
	ctx.Scopes = ctx.Scopes[:len(ctx.Scopes)-1]
}

func (ctx *NodeResolutionContext) CurrentScope() *Scope {
	if len(ctx.Scopes) == 0 {
		panic("no scopes in context")
	}

	return &ctx.Scopes[len(ctx.Scopes)-1]
}

func (ctx *NodeResolutionContext) Bind(fields []fields.Field) {
	scope := ctx.CurrentScope()
	scope.fields = append(scope.fields, fields...)
}

func (ctx *NodeResolutionContext) Lookup(relationName, name string) (fields.Field, error) {
	qualifiedSearchName := fmt.Sprintf("%q", name)
	if relationName != "" {
		qualifiedSearchName = fmt.Sprintf("%q.%q", relationName, name)
	}

	for i := len(ctx.Scopes) - 1; i >= 0; i-- {
		scope := ctx.Scopes[i]

		var candidates []fields.Field
		for _, field := range scope.fields {
			if (field.RelationName() == relationName || relationName == "") && field.Name() == name {
				candidates = append(candidates, field)
			}
		}

		if len(candidates) == 1 {
			return candidates[0], nil
		}

		if len(candidates) > 1 {
			return fields.Field{}, fmt.Errorf("ambiguous field %s", qualifiedSearchName)
		}
	}

	return fields.Field{}, fmt.Errorf("unknown field %s", qualifiedSearchName)
}

//
//

type OptimizationContext struct {
	catalog     CatalogSet
	outerFields []fields.Field
}

func NewOptimizationContext(catalog CatalogSet) OptimizationContext {
	return OptimizationContext{
		catalog: catalog,
	}
}

func (c OptimizationContext) Catalog() CatalogSet {
	return c.catalog
}

func (c OptimizationContext) OuterFields() []fields.Field {
	return c.outerFields
}

func (c OptimizationContext) AddOuterFields(fields []fields.Field) OptimizationContext {
	c.outerFields = append(c.outerFields, fields...)
	return c
}

//
//

type ExecutionContext struct {
	catalog  CatalogSet
	debug    bool
	outerRow rows.Row
}

var EmptyExecutionContext = NewExecutionContext(NewCatalogEmptySet())

func NewExecutionContext(catalog CatalogSet) ExecutionContext {
	return ExecutionContext{
		catalog: catalog,
	}
}

func (c ExecutionContext) OptimizationContext() OptimizationContext {
	return NewOptimizationContext(c.catalog)
}

func (c ExecutionContext) Catalog() CatalogSet {
	return c.catalog
}

func (c ExecutionContext) OuterRow() rows.Row {
	return c.outerRow
}

func (c ExecutionContext) WithDebug() ExecutionContext {
	c.debug = true
	return c
}

func (c ExecutionContext) AddOuterRow(row rows.Row) ExecutionContext {
	c.outerRow = rows.CombineRows(c.outerRow, row)
	return c
}

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
