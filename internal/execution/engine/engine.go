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
	catalog impls.CatalogSet
}

func NewDefaultEngine() *Engine {
	return NewEngine(impls.NewCatalogSet(
		catalog.NewCatalog[impls.Table](),
		catalog.NewCatalog[impls.Sequence](),
		catalog.NewCatalogWithEntries[impls.Function](functions.DefaultFunctions()),
		catalog.NewCatalogWithEntries[impls.Aggregate](aggregates.DefaultAggregates()),
	))
}

func NewEngine(catalog impls.CatalogSet) *Engine {
	return &Engine{
		catalog: catalog,
	}
}

func (e *Engine) Query(request protocol.Request, responseWriter protocol.ResponseWriter) {
	query, err := parsing.Parse(e.catalog.Tables, lexing.Lex(request.Query))
	if err != nil {
		responseWriter.Error(fmt.Errorf("failed to parse query: %s", err))
		return
	}

	executionContext := impls.NewContext(e.catalog)
	if request.Debug {
		executionContext = executionContext.WithDebug()
	}

	query.Execute(executionContext, responseWriter)
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
