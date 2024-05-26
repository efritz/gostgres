package engine

import (
	"fmt"

	"github.com/efritz/gostgres/internal/catalog"
	"github.com/efritz/gostgres/internal/catalog/aggregates"
	"github.com/efritz/gostgres/internal/catalog/functions"
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/lexing"
	"github.com/efritz/gostgres/internal/syntax/parsing"
	"github.com/efritz/gostgres/internal/types"
)

type Engine struct {
	tables     *catalog.Catalog[types.Table]
	sequences  *catalog.Catalog[types.Sequence]
	functions  *catalog.Catalog[types.Function]
	aggregates *catalog.Catalog[types.Aggregate]
}

func NewDefaultEngine() *Engine {
	return NewEngine(
		catalog.NewCatalog[types.Table](),
		catalog.NewCatalog[types.Sequence](),
		catalog.NewCatalogWithEntries[types.Function](functions.DefaultFunctions()),
		catalog.NewCatalogWithEntries[types.Aggregate](aggregates.DefaultAggregates()),
	)
}

func NewEngine(
	tables *catalog.Catalog[types.Table],
	sequences *catalog.Catalog[types.Sequence],
	functions *catalog.Catalog[types.Function],
	aggregates *catalog.Catalog[types.Aggregate],
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

	ctx := types.NewContext(
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
