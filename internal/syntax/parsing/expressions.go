package parsing

import (
	"fmt"
	"strconv"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

func (p *parser) parseExpression(precedence Precedence) (expressions.Expression, error) {
	expression, err := p.parseExpressionPrefix()
	if err != nil {
		return nil, err
	}

	return p.parseExpressionSuffix(expression, precedence)
}

func (p *parser) parseExpressionPrefix() (expressions.Expression, error) {
	current := p.advance()
	parseFunc, ok := p.prefixParsers[current.Type]
	if !ok {
		return nil, fmt.Errorf("expected expression (near %s)", current.Text)
	}

	return parseFunc(current)
}

func (p *parser) parseExpressionSuffix(expression expressions.Expression, precedence Precedence) (_ expressions.Expression, err error) {
	for {
		tokenType := p.current().Type
		parseFunc, ok := p.infixParsers[tokenType]
		if !ok {
			break
		}
		tokenPrecedence, ok := precedenceMap[tokenType]
		if !ok || tokenPrecedence < precedence {
			break
		}

		expression, err = parseFunc(expression, p.advance())
		if err != nil {
			return nil, err
		}
	}

	return expression, nil
}

// namedExpression := ident
//                  | ident `.` ident
func (p *parser) parseNamedExpression(token tokens.Token) (expressions.Expression, error) {
	if !p.advanceIf(isType(tokens.TokenTypeDot)) {
		return expressions.NewNamed(shared.NewField("", token.Text, shared.TypeKindAny, false)), nil
	}

	qualifiedNameToken, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
	if err != nil {
		return nil, err
	}

	return expressions.NewNamed(shared.NewField(token.Text, qualifiedNameToken.Text, shared.TypeKindAny, false)), nil
}

// numericLiteralExpression := number
func (p *parser) parseNumericLiteralExpression(token tokens.Token) (expressions.Expression, error) {
	value, err := strconv.Atoi(token.Text)
	if err != nil {
		return nil, err
	}

	return expressions.NewConstant(value), nil
}

// numericLiteralExpression := string
func (p *parser) parseStringLiteralExpression(token tokens.Token) (expressions.Expression, error) {
	return expressions.NewConstant(token.Text), nil
}

// numericLiteralExpression := true | false
func (p *parser) parseBooleanLiteralExpression(token tokens.Token) (expressions.Expression, error) {
	return expressions.NewConstant(token.Type == tokens.TokenTypeTrue), nil
}

// numericLiteralExpression := null
func (p *parser) parseNullLiteralExpression(token tokens.Token) (expressions.Expression, error) {
	return expressions.NewConstant(nil), nil
}

// parenthesizedExpression := `(` expression `)`
func (p *parser) parseParenthesizedExpression(token tokens.Token) (expressions.Expression, error) {
	inner, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeRightParen)); err != nil {
		return nil, err
	}

	return inner, nil
}

func (p *parser) parseUnary(factory unaryExpressionParserFunc) prefixParserFunc {
	return func(token tokens.Token) (expressions.Expression, error) {
		expression, err := p.parseExpression(PrecedenceUnary)
		if err != nil {
			return nil, err
		}

		return factory(expression), nil
	}
}

func (p *parser) parseBinary(precedence Precedence, factory binaryExpressionParserFunc) infixParserFunc {
	return func(left expressions.Expression, token tokens.Token) (expressions.Expression, error) {
		right, err := p.parseExpression(precedence)
		if err != nil {
			return nil, err
		}

		return factory(left, right), nil
	}
}

func (p *parser) parsePostfix(precedence Precedence, factory unaryExpressionParserFunc) infixParserFunc {
	return func(left expressions.Expression, token tokens.Token) (expressions.Expression, error) {
		return factory(left), nil
	}
}

func (p *parser) parseBetween(factory ternaryExpressionParserFunc) infixParserFunc {
	return func(left expressions.Expression, token tokens.Token) (expressions.Expression, error) {
		middle, err := p.parseExpressionPrefix()
		if err != nil {
			return nil, err
		}

		if _, err := p.mustAdvance(isType(tokens.TokenTypeAnd)); err != nil {
			return nil, err
		}

		right, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		return factory(left, middle, right), nil
	}
}
