package parsing

import (
	"fmt"

	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/syntax/tokens"
	"github.com/efritz/gostgres/internal/table"
)

func Parse(tokenStream []tokens.Token, tables map[string]*table.Table) (queries.Node, error) {
	parser := newParser(tokenStream, tables)
	statement, err := parser.parseStatement()
	if err != nil {
		return nil, err
	}

	_ = parser.advanceIf(isType(tokens.TokenTypeSemicolon))
	if parser.cursor < len(parser.tokens) {
		return nil, fmt.Errorf("unexpected tokens at end of statement (near %s)", parser.tokens[parser.cursor].Text)
	}

	return statement, nil
}
