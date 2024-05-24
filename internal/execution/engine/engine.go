package engine

import (
	"fmt"

	"github.com/efritz/gostgres/internal/catalog"
	"github.com/efritz/gostgres/internal/catalog/aggregates"
	"github.com/efritz/gostgres/internal/catalog/functions"
	"github.com/efritz/gostgres/internal/catalog/sequence"
	"github.com/efritz/gostgres/internal/catalog/table"
	"github.com/efritz/gostgres/internal/execution"
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/lexing"
	"github.com/efritz/gostgres/internal/syntax/parsing"
)

type Engine struct {
	tables     *catalog.Catalog[*table.Table]
	sequences  *catalog.Catalog[*sequence.Sequence]
	functions  *catalog.Catalog[functions.Function]
	aggregates *catalog.Catalog[aggregates.Aggregate]
}

func NewDefaultEngine() *Engine {
	return NewEngine(
		catalog.NewCatalog[*table.Table](),
		catalog.NewCatalog[*sequence.Sequence](),
		catalog.NewCatalogWithEntries[functions.Function](functions.DefaultFunctions()),
		catalog.NewCatalogWithEntries[aggregates.Aggregate](aggregates.DefaultAggregates()),
	)
}

func NewEngine(
	tables *catalog.Catalog[*table.Table],
	sequences *catalog.Catalog[*sequence.Sequence],
	functions *catalog.Catalog[functions.Function],
	aggregates *catalog.Catalog[aggregates.Aggregate],
) *Engine {
	return &Engine{
		tables:     tables,
		sequences:  sequences,
		functions:  functions,
		aggregates: aggregates,
	}
}

func (e *Engine) Query(input string) (shared.Rows, error) {
	query, err := parsing.Parse(lexing.Lex(input), e.tables)
	if err != nil {
		return shared.Rows{}, fmt.Errorf("failed to parse query: %s", err)
	}

	ctx := execution.NewContext(
		e.tables,
		e.sequences,
		e.functions,
		e.aggregates,
	)

	collector := protocol.NewRowCollector()
	query.Execute(ctx, collector)
	rows, err := collector.Rows()
	if err != nil {
		return shared.Rows{}, fmt.Errorf("failed to execute query %q: %s", input, err)
	}

	return rows, nil
}
