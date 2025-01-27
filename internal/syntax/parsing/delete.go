package parsing

import (
	"github.com/efritz/gostgres/internal/syntax/ast"
	"github.com/efritz/gostgres/internal/syntax/ast/mutation"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// deleteTail := `FROM` table deleteUsing where returning
func (p *parser) parseDelete(token tokens.Token) (ast.BuilderResolver, error) {
	if _, err := p.mustAdvance(isType(tokens.TokenTypeFrom)); err != nil {
		return nil, err
	}

	tableDescription, err := p.parseTable()
	if err != nil {
		return nil, err
	}

	usingExpressions, err := p.parseDeleteUsing()
	if err != nil {
		return nil, err
	}

	whereExpression, _, err := p.parseWhere()
	if err != nil {
		return nil, err
	}

	returningExpressions, err := p.parseReturning()
	if err != nil {
		return nil, err
	}

	return &mutation.DeleteBuilder{
		Target:    tableDescription,
		Using:     usingExpressions,
		Where:     whereExpression,
		Returning: returningExpressions,
	}, nil
}

// deleteUsing := `USING` tableExpressions
func (p *parser) parseDeleteUsing() ([]*ast.TableExpression, error) {
	if !p.advanceIf(isType(tokens.TokenTypeUsing)) {
		return nil, nil
	}

	return p.parseTableExpressions()
}
