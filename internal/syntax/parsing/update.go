package parsing

import (
	"github.com/efritz/gostgres/internal/syntax/ast"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// updateTail := table `SET` ( setExpression [, ...] ) [ `FROM` tableExpressions ] where returning
func (p *parser) parseUpdate(token tokens.Token) (ast.BuilderResolver, error) {
	tableDescription, err := p.parseTable()
	if err != nil {
		return nil, err
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeSet)); err != nil {
		return nil, err
	}

	setExpressions, err := parseCommaSeparatedList(p, p.parseSetExpression)
	if err != nil {
		return nil, err
	}

	var fromExpressions []*ast.TableExpression
	if p.advanceIf(isType(tokens.TokenTypeFrom)) {
		fromExpressions, err = p.parseTableExpressions()
		if err != nil {
			return nil, err
		}
	}

	whereExpression, _, err := p.parseWhere()
	if err != nil {
		return nil, err
	}

	returningExpressions, err := p.parseReturning()
	if err != nil {
		return nil, err
	}

	return &ast.UpdateBuilder{
		Target:    tableDescription,
		Updates:   setExpressions,
		From:      fromExpressions,
		Where:     whereExpression,
		Returning: returningExpressions,
	}, nil
}

// setExpression := ident `=` expression
func (p *parser) parseSetExpression() (ast.SetExpression, error) {
	if p.advanceIf(isType(tokens.TokenTypeLeftParen)) {
		// TODO - support sub-select
		// TODO - support row values
		panic("Multi-column sets unimplemented")
	}

	name, err := p.parseIdent()
	if err != nil {
		return ast.SetExpression{}, err
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeEquals)); err != nil {
		return ast.SetExpression{}, err
	}

	expr, err := p.parseRootExpression()
	if err != nil {
		return ast.SetExpression{}, err
	}

	return ast.SetExpression{
		Name:       name,
		Expression: expr,
	}, nil
}
