package parsing

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/explain"
	"github.com/efritz/gostgres/internal/syntax/ast"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

func (p *parser) initDDLParsers() {
	p.ddlParsers = ddlParsers{
		tokens.TokenTypeCreate: p.parseCreate,
		tokens.TokenTypeAlter:  p.parseAlter,
	}
}

func (p *parser) initStatementParsers() {
	p.explainableParsers = explainableParsers{
		tokens.TokenTypeSelect: func(_ tokens.Token) (ast.Builder, error) { return p.parseSelect() },
		tokens.TokenTypeInsert: p.parseInsert,
		tokens.TokenTypeUpdate: p.parseUpdate,
		tokens.TokenTypeDelete: p.parseDelete,
	}
}

// statement := ddlStatement | ( [ `EXPLAIN` ] explainableStatement )
// ddlStatement := ( `CREATE` createTail ) | ( `ALTER` alterTail )
// explainableStatement := ( `SELECT` selectTail ) | ( `INSERT` insertTail ) | ( `UPDATE` updateTail ) | ( `DELETE` deleteTail )
func (p *parser) parseStatement(ctx ast.BuildContext) (Query, error) {
	for tokenType, parser := range p.ddlParsers {
		token := p.current()
		if p.advanceIf(isType(tokenType)) {
			return parser(token)
		}
	}

	isExplain := false
	if p.advanceIf(isType(tokens.TokenTypeExplain)) {
		isExplain = true
	}

	for tokenType, parser := range p.explainableParsers {
		token := p.current()
		if p.advanceIf(isType(tokenType)) {
			builder, err := parser(token)
			if err != nil {
				return nil, err
			}

			node, err := builder.Build(ctx)
			if err != nil {
				return nil, err
			}

			if isExplain {
				node = explain.NewExplain(node)
			}

			return queries.NewQuery(node), nil
		}
	}

	return nil, fmt.Errorf("expected start of statement (near %s)", p.current().Text)
}
