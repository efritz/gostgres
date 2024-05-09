package engine

import (
	"fmt"

	"github.com/efritz/gostgres/internal/functions"
	"github.com/efritz/gostgres/internal/protocol"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/sequence"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/lexing"
	"github.com/efritz/gostgres/internal/syntax/parsing"
	"github.com/efritz/gostgres/internal/table"
)

type Engine struct {
	tables    *table.Tablespace
	sequences *sequence.Sequencespace
	functions *functions.Functionspace
}

func NewEngine(tables *table.Tablespace, sequences *sequence.Sequencespace, functions *functions.Functionspace) *Engine {
	return &Engine{
		tables:    tables,
		sequences: sequences,
		functions: functions,
	}
}

func (e *Engine) Query(input string) (shared.Rows, error) {
	query, err := parsing.Parse(lexing.Lex(input), e.tables)
	if err != nil {
		return shared.Rows{}, fmt.Errorf("failed to parse query: %s", err)
	}

	collector := protocol.NewRowCollector()
	query.Execute(queries.NewContext(e.tables, e.sequences, e.functions), collector)
	rows, err := collector.Rows()
	if err != nil {
		return shared.Rows{}, fmt.Errorf("failed to execute query: %s", err)
	}

	return rows, nil
}
