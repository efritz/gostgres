package parsing

import (
	"github.com/efritz/gostgres/internal/syntax/ast"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// analyzeTail := ( ident [, ...] )
func (p *parser) parseAnalyze(token tokens.Token) (ast.BuilderResolver, error) {
	if p.peek(0).Type != tokens.TokenTypeIdent {
		return &ast.AnalyzeBuilder{}, nil
	}

	tableNames, err := parseCommaSeparatedList(p, p.parseIdent)
	if err != nil {
		return nil, err
	}

	return &ast.AnalyzeBuilder{TableNames: tableNames}, nil
}
