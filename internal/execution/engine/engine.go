package engine

import (
	"fmt"

	"github.com/efritz/gostgres/internal/catalog"
	"github.com/efritz/gostgres/internal/catalog/aggregates"
	"github.com/efritz/gostgres/internal/catalog/functions"
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/syntax/lexing"
	"github.com/efritz/gostgres/internal/syntax/parsing"
)

type Engine struct {
	tables     *catalog.Catalog[impls.Table]
	sequences  *catalog.Catalog[impls.Sequence]
	functions  *catalog.Catalog[impls.Function]
	aggregates *catalog.Catalog[impls.Aggregate]
}

func NewDefaultEngine() *Engine {
	return NewEngine(
		catalog.NewCatalog[impls.Table](),
		catalog.NewCatalog[impls.Sequence](),
		catalog.NewCatalogWithEntries[impls.Function](functions.DefaultFunctions()),
		catalog.NewCatalogWithEntries[impls.Aggregate](aggregates.DefaultAggregates()),
	)
}

func NewEngine(
	tables *catalog.Catalog[impls.Table],
	sequences *catalog.Catalog[impls.Sequence],
	functions *catalog.Catalog[impls.Function],
	aggregates *catalog.Catalog[impls.Aggregate],
) *Engine {
	return &Engine{
		tables:     tables,
		sequences:  sequences,
		functions:  functions,
		aggregates: aggregates,
	}
}

func (e *Engine) Query(request protocol.Request, responseWriter protocol.ResponseWriter) {
	query, err := parsing.Parse(lexing.Lex(request.Query), e.tables)
	if err != nil {
		responseWriter.Error(fmt.Errorf("failed to parse query: %s", err))
		return
	}

	ctx := impls.NewContext(
		e.tables,
		e.sequences,
		e.functions,
		e.aggregates,
	)
	if request.Debug {
		ctx = ctx.WithDebug()
	}

	query.Execute(ctx, responseWriter)
}

func (e *Engine) QueryRows(request protocol.Request) (rows.Rows, error) {
	collector := protocol.NewRowCollector()
	e.Query(request, collector)
	collectedRows, err := collector.Rows()
	if err != nil {
		return rows.Rows{}, fmt.Errorf("failed to execute query %q: %s", request.Query, err)
	}

	return collectedRows, nil
}

func (e *Engine) QueryError(request protocol.Request) error {
	collector := protocol.NewRowCollector()
	e.Query(request, collector)
	_, err := collector.Rows()
	return err
}
