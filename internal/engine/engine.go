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
	tables     *table.Tablespace
	sequences  *sequence.Sequencespace
	functions  *functions.Functionspace
	aggregates *aggregates.Aggregatespace
}

func NewDefaultEngine() *Engine {
	return NewEngine(
		table.NewTablespace(),
		sequence.NewSequencespace(),
		functions.NewDefaultFunctionspace(),
		aggregates.NewDefaultAggregatespace(),
	)
}

func NewEngine(
	tables *table.Tablespace,
	sequences *sequence.Sequencespace,
	functions *functions.Functionspace,
	aggregates *aggregates.Aggregatespace,
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
