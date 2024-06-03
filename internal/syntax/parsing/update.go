package parsing

import (
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// updateTail := table `SET` ( setExpression [, ...] ) [ `FROM` tableExpressions ] where returning
func (p *parser) parseUpdate(token tokens.Token) (Builder, error) {
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

	var fromExpressions []TableExpressionDescription
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

	returningExpressions, err := p.parseReturning(tableDescription.name)
	if err != nil {
		return nil, err
	}

	return &UpdateBuilder{
		tableDescription:     tableDescription,
		setExpressions:       setExpressions,
		fromExpressions:      fromExpressions,
		whereExpression:      whereExpression,
		returningExpressions: returningExpressions,
	}, nil
}

// setExpression := ident `=` expression
func (p *parser) parseSetExpression() (SetExpression, error) {
	if p.advanceIf(isType(tokens.TokenTypeLeftParen)) {
		// TODO - support sub-select
		// TODO - support row values
		panic("Multi-column sets unimplemented")
	}

	name, err := p.parseIdent()
	if err != nil {
		return SetExpression{}, err
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeEquals)); err != nil {
		return SetExpression{}, err
	}

	expr, err := p.parseRootExpression()
	if err != nil {
		return SetExpression{}, err
	}

	return SetExpression{
		Name:       name,
		Expression: expr,
	}, nil
}
