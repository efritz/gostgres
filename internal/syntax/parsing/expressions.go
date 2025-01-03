package parsing

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/types"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// expression := ...
func (p *parser) parseRootExpression() (impls.Expression, error) {
	return p.parseExpression(0)
}

// expressionWithDirection := expression [ `ASC` | `DESC` ]
func (p *parser) parseExpressionWithDirection() (impls.ExpressionWithDirection, error) {
	expression, err := p.parseRootExpression()
	if err != nil {
		return impls.ExpressionWithDirection{}, err
	}

	reverse := false
	if p.advanceIf(isType(tokens.TokenTypeAscending)) {
		// no-op
	} else if p.advanceIf(isType(tokens.TokenTypeDescending)) {
		reverse = true
	}

	return impls.ExpressionWithDirection{
		Expression: expression,
		Reverse:    reverse,
	}, nil
}

func (p *parser) parseExpression(precedence Precedence) (impls.Expression, error) {
	expression, err := p.parseExpressionPrefix()
	if err != nil {
		return nil, err
	}

	return p.parseExpressionSuffix(expression, precedence)
}

func (p *parser) initExpressionPrefixParsers() {
	p.prefixParsers = prefixParsers{
		tokens.TokenTypeIdent:     p.parseNamedExpression,
		tokens.TokenTypeNumber:    p.parseNumericLiteralExpression,
		tokens.TokenTypeString:    p.parseStringLiteralExpression,
		tokens.TokenTypeFalse:     p.parseBooleanLiteralExpression,
		tokens.TokenTypeNot:       p.parseUnary(expressions.NewNot),
		tokens.TokenTypeNull:      p.parseNullLiteralExpression,
		tokens.TokenTypeTrue:      p.parseBooleanLiteralExpression,
		tokens.TokenTypeMinus:     p.parseUnary(expressions.NewUnaryMinus),
		tokens.TokenTypeLeftParen: p.parseParenthesizedExpression,
		tokens.TokenTypePlus:      p.parseUnary(expressions.NewUnaryPlus),
		tokens.TokenTypeExists:    p.parseExistsExpression,
	}
}

func (p *parser) parseExpressionPrefix() (impls.Expression, error) {
	current := p.advance()
	parseFunc, ok := p.prefixParsers[current.Type]
	if !ok {
		return nil, fmt.Errorf("expected expression (near %s)", current.Text)
	}

	return parseFunc(current)
}

func (p *parser) initExpressionInfixParsers() {
	p.infixParsers = infixParsers{
		tokens.TokenTypeAnd:                 p.parseBinary(PrecedenceConditionalAnd, expressions.NewAnd),
		tokens.TokenTypeOr:                  p.parseBinary(PrecedenceConditionalOr, expressions.NewOr),
		tokens.TokenTypeMinus:               p.parseBinary(PrecedenceAdditive, expressions.NewSubtraction),
		tokens.TokenTypeAsterisk:            p.parseBinary(PrecedenceMultiplicative, expressions.NewMultiplication),
		tokens.TokenTypeSlash:               p.parseBinary(PrecedenceMultiplicative, expressions.NewDivision),
		tokens.TokenTypePlus:                p.parseBinary(PrecedenceAdditive, expressions.NewAddition),
		tokens.TokenTypeLessThan:            p.parseBinary(PrecedenceComparison, expressions.NewLessThan),
		tokens.TokenTypeEquals:              p.parseBinary(PrecedenceEquality, expressions.NewEquals),
		tokens.TokenTypeGreaterThan:         p.parseBinary(PrecedenceComparison, expressions.NewGreaterThan),
		tokens.TokenTypeLessThanOrEqual:     p.parseBinary(PrecedenceComparison, expressions.NewLessThanEquals),
		tokens.TokenTypeNotEquals:           negate(p.parseBinary(PrecedenceEquality, expressions.NewEquals)),
		tokens.TokenTypeGreaterThanOrEqual:  p.parseBinary(PrecedenceComparison, expressions.NewGreaterThanEquals),
		tokens.TokenTypeIsTrue:              p.parsePostfix(PrecedencePostfix, expressions.NewIsTrue),
		tokens.TokenTypeIsNotTrue:           negate(p.parsePostfix(PrecedencePostfix, expressions.NewIsTrue)),
		tokens.TokenTypeIsFalse:             p.parsePostfix(PrecedencePostfix, expressions.NewIsFalse),
		tokens.TokenTypeIsNotFalse:          negate(p.parsePostfix(PrecedencePostfix, expressions.NewIsFalse)),
		tokens.TokenTypeIsNull:              p.parsePostfix(PrecedencePostfix, expressions.NewIsNull),
		tokens.TokenTypeIsNotNull:           negate(p.parsePostfix(PrecedencePostfix, expressions.NewIsNull)),
		tokens.TokenTypeIsUnknown:           p.parsePostfix(PrecedencePostfix, expressions.NewIsUnknown),
		tokens.TokenTypeIsNotUnknown:        negate(p.parsePostfix(PrecedencePostfix, expressions.NewIsUnknown)),
		tokens.TokenTypeConcat:              p.parseBinary(PrecedenceGenericOperator, expressions.NewConcat),
		tokens.TokenTypeIsDistinctFrom:      p.parseBinary(PrecedenceIs, expressions.NewIsDistinctFrom),
		tokens.TokenTypeIsNotDistinctFrom:   negate(p.parseBinary(PrecedenceIs, expressions.NewIsDistinctFrom)),
		tokens.TokenTypeLike:                p.parseBinary(PrecedenceLike, expressions.NewLike),
		tokens.TokenTypeNotLike:             negate(p.parseBinary(PrecedenceLike, expressions.NewLike)),
		tokens.TokenTypeILike:               p.parseBinary(PrecedenceLike, expressions.NewILike),
		tokens.TokenTypeNotILike:            negate(p.parseBinary(PrecedenceLike, expressions.NewILike)),
		tokens.TokenTypeBetween:             p.parseBetween(expressions.NewBetween),
		tokens.TokenTypeNotBetween:          negate(p.parseBetween(expressions.NewBetween)),
		tokens.TokenTypeBetweenSymmetric:    p.parseBetween(expressions.NewBetweenSymmetric),
		tokens.TokenTypeNotBetweenSymmetric: negate(p.parseBetween(expressions.NewBetweenSymmetric)),
	}
}

func (p *parser) parseExpressionSuffix(expression impls.Expression, precedence Precedence) (_ impls.Expression, err error) {
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

func (p *parser) parseStringLiteralExpression(token tokens.Token) (impls.Expression, error) {
	return expressions.NewConstant(token.Text), nil
}

func (p *parser) parseBooleanLiteralExpression(token tokens.Token) (impls.Expression, error) {
	return expressions.NewConstant(token.Type == tokens.TokenTypeTrue), nil
}

func (p *parser) parseNullLiteralExpression(token tokens.Token) (impls.Expression, error) {
	return expressions.NewConstant(nil), nil
}

func (p *parser) parseNumericLiteralExpression(token tokens.Token) (impls.Expression, error) {
	if p.advanceIf(isType(tokens.TokenTypeDot)) {
		fractionalPart, err := p.mustAdvance(isType(tokens.TokenTypeNumber))
		if err != nil {
			return nil, err
		}

		value, err := strconv.ParseFloat(token.Text+"."+fractionalPart.Text, 64)
		if err != nil {
			return nil, err
		}

		return expressions.NewConstant(float64(value)), err
	}

	value, err := strconv.Atoi(token.Text)
	if err != nil {
		return nil, err
	}

	return expressions.NewConstant(int32(value)), nil
}

// parenthesizedExpressionTail := expression `)`
func (p *parser) parseParenthesizedExpression(token tokens.Token) (impls.Expression, error) {
	inner, err := p.parseRootExpression()
	if err != nil {
		return nil, err
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeRightParen)); err != nil {
		return nil, err
	}

	return inner, nil
}

// namedExpressionTail := ( `.` ident ) | ( functionInvocationTail ) | <empty>
func (p *parser) parseNamedExpression(token tokens.Token) (impls.Expression, error) {
	if p.advanceIf(isType(tokens.TokenTypeDot)) {
		qualifiedNameToken, err := p.parseIdent()
		if err != nil {
			return nil, err
		}

		return expressions.NewNamed(fields.NewField(token.Text, qualifiedNameToken, types.TypeAny, fields.NonInternalField)), nil
	}

	if p.peek(0).Type == tokens.TokenTypeLeftParen {
		return p.parseFunctionInvocationTail(token)
	}

	return expressions.NewNamed(fields.NewField("", token.Text, types.TypeAny, fields.NonInternalField)), nil
}

// functionInvocationTail := `(` [ expression [, ...] ] `)`
func (p *parser) parseFunctionInvocationTail(token tokens.Token) (impls.Expression, error) {
	if next := p.peek(0); next.Type != tokens.TokenTypeLeftParen {
		return nil, fmt.Errorf("expected left paren (near %s)", token.Text)
	}

	// Handle special case for COUNT(*) -> COUNT(1)
	if strings.ToLower(token.Text) == "count" && p.advanceIf(isType(tokens.TokenTypeLeftParen), isType(tokens.TokenTypeAsterisk), isType(tokens.TokenTypeRightParen)) {
		return expressions.NewFunction(token.Text, []impls.Expression{expressions.NewConstant(1)}), nil
	}

	args, err := parseParenthesizedCommaSeparatedList(p, false, true, p.parseRootExpression)
	if err != nil {
		return nil, err
	}

	return expressions.NewFunction(token.Text, args), nil
}

type unaryExpressionParserFunc func(expression impls.Expression) impls.Expression
type binaryExpressionParserFunc func(left, right impls.Expression) impls.Expression
type ternaryExpressionParserFunc func(left, middle, right impls.Expression) impls.Expression

func (p *parser) parseUnary(factory unaryExpressionParserFunc) prefixParserFunc {
	return func(token tokens.Token) (impls.Expression, error) {
		expression, err := p.parseExpression(PrecedenceUnary)
		if err != nil {
			return nil, err
		}

		return factory(expression), nil
	}
}

func (p *parser) parseBinary(precedence Precedence, factory binaryExpressionParserFunc) infixParserFunc {
	return func(left impls.Expression, token tokens.Token) (impls.Expression, error) {
		right, err := p.parseExpression(precedence)
		if err != nil {
			return nil, err
		}

		return factory(left, right), nil
	}
}

func (p *parser) parsePostfix(_ Precedence, factory unaryExpressionParserFunc) infixParserFunc {
	return func(left impls.Expression, token tokens.Token) (impls.Expression, error) {
		return factory(left), nil
	}
}

func (p *parser) parseBetween(factory ternaryExpressionParserFunc) infixParserFunc {
	return func(left impls.Expression, token tokens.Token) (impls.Expression, error) {
		middle, err := p.parseExpressionPrefix()
		if err != nil {
			return nil, err
		}

		if _, err := p.mustAdvance(isType(tokens.TokenTypeAnd)); err != nil {
			return nil, err
		}

		right, err := p.parseRootExpression()
		if err != nil {
			return nil, err
		}

		return factory(left, middle, right), nil
	}
}

func negate(parserFunc infixParserFunc) infixParserFunc {
	return func(left impls.Expression, token tokens.Token) (impls.Expression, error) {
		expression, err := parserFunc(left, token)
		if err != nil {
			return nil, err
		}

		return expressions.NewNot(expression), nil
	}
}

// exists := `EXISTS` `(` subquery `)`
func (p *parser) parseExistsExpression(token tokens.Token) (impls.Expression, error) {
	if _, err := p.mustAdvance(isType(tokens.TokenTypeLeftParen)); err != nil {
		return nil, err
	}

	subquery, err := p.parseSelectOrValues()
	if err != nil {
		return nil, err
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeRightParen)); err != nil {
		return nil, err
	}

	return expressions.NewExistsSubqueryExpression(subquery), nil
}

// IN/NOT IN (subquery) => = ANY/!= ANY
// op ANY/SOME (subquery)
// op ALL (subquery)

// TODO
// row_constructor    {IN / NOT IN}      (subquery)
// row_constructor op {ALL / ANY / SOME} (subquery)
// row_constructor op ALL                (subquery)
// row_constructor op (single-row subquery)
