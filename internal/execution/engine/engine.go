package engine

import (
	"fmt"

	"github.com/efritz/gostgres/internal/catalog"
	"github.com/efritz/gostgres/internal/catalog/aggregates"
	"github.com/efritz/gostgres/internal/catalog/functions"
	"github.com/efritz/gostgres/internal/execution/engine/protocol"
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

func (e *Engine) Query(input string, debug bool) (rows.Rows, error) {
	query, err := parsing.Parse(lexing.Lex(input), e.tables)
	if err != nil {
		return rows.Rows{}, fmt.Errorf("failed to parse query: %s", err)
	}

	ctx := impls.NewContext(
		e.tables,
		e.sequences,
		e.functions,
		e.aggregates,
	)
	if debug {
		ctx = ctx.WithDebug()
	}

	collector := protocol.NewRowCollector()
	query.Execute(ctx, collector)
	collectedRows, err := collector.Rows()
	if err != nil {
		return rows.Rows{}, fmt.Errorf("failed to execute query %q: %s", input, err)
	}

	return collectedRows, nil
}
