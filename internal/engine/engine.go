package engine

import (
	"fmt"

	"github.com/efritz/gostgres/internal/aggregates"
	"github.com/efritz/gostgres/internal/eval"
	"github.com/efritz/gostgres/internal/functions"
	"github.com/efritz/gostgres/internal/protocol"
	"github.com/efritz/gostgres/internal/sequence"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/lexing"
	"github.com/efritz/gostgres/internal/syntax/parsing"
	"github.com/efritz/gostgres/internal/table"
)

type Engine struct {
	tables     *eval.Catalog[*table.Table]
	sequences  *eval.Catalog[*sequence.Sequence]
	functions  *eval.Catalog[functions.Function]
	aggregates *eval.Catalog[aggregates.Aggregate]
}

func NewDefaultEngine() *Engine {
	return NewEngine(
		eval.NewCatalog[*table.Table](),
		eval.NewCatalog[*sequence.Sequence](),
		eval.NewCatalogWithEntries[functions.Function](functions.DefaultFunctions()),
		eval.NewCatalogWithEntries[aggregates.Aggregate](aggregates.DefaultAggregates()),
	)
}

func NewEngine(
	tables *eval.Catalog[*table.Table],
	sequences *eval.Catalog[*sequence.Sequence],
	functions *eval.Catalog[functions.Function],
	aggregates *eval.Catalog[aggregates.Aggregate],
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

	ctx := eval.NewContext(
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
