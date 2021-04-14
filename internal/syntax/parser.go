package syntax

import (
	"fmt"
	"strconv"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/relations"
)

func Parse(tokens []Token, builtins map[string]relations.Relation) (relations.Relation, error) {
	parser := &parser{
		tokens:   tokens,
		builtins: builtins,
	}
	parser.init()

	statement, err := parser.parseStatement()
	if err != nil {
		return nil, err
	}

	_ = parser.advanceIf(isType(TokenTypeSemicolon))
	if parser.cursor < len(parser.tokens) {
		return nil, fmt.Errorf("unexpected tokens at end of statement (near %s)", parser.tokens[parser.cursor].Text)
	}

	return statement, nil
}

type parser struct {
	tokens        []Token
	cursor        int
	builtins      map[string]relations.Relation
	prefixParsers map[TokenType]prefixParserFunc
	infixParsers  map[TokenType]infixParserFunc
}

type tokenFilterFunc func(t Token) bool
type prefixParserFunc func(token Token) (expressions.Expression, error)
type infixParserFunc func(left expressions.Expression, token Token) (expressions.Expression, error)
type unaryExpressionParserFunc func(expression expressions.Expression) expressions.Expression
type binaryExpressionParserFunc func(left, right expressions.Expression) expressions.Expression

type Precedence int

const (
	PrecedenceUnknown Precedence = iota
	PrecedenceConditionalOr
	PrecedenceConditionalAnd
	PrecedenceEquality
	PrecedenceComparison
	PrecedenceIs
	PrecedenceAdditive
	PrecedenceMultiplicative
	PrecedenceUnary
	PrecedenceAny
)

var precedenceMap = map[TokenType]Precedence{
	TokenTypeAnd:                PrecedenceConditionalAnd,
	TokenTypeIs:                 PrecedenceIs,
	TokenTypeOr:                 PrecedenceConditionalOr,
	TokenTypeMinus:              PrecedenceAdditive,
	TokenTypeAsterisk:           PrecedenceMultiplicative,
	TokenTypeSlash:              PrecedenceMultiplicative,
	TokenTypePlus:               PrecedenceAdditive,
	TokenTypeLessThan:           PrecedenceComparison,
	TokenTypeEquals:             PrecedenceEquality,
	TokenTypeGreaterThan:        PrecedenceComparison,
	TokenTypeLessThanOrEqual:    PrecedenceComparison,
	TokenTypeNotEquals:          PrecedenceEquality,
	TokenTypeGreaterThanOrEqual: PrecedenceComparison,
}

func (p *parser) init() {
	p.prefixParsers = map[TokenType]prefixParserFunc{
		TokenTypeIdent:     p.parseNamedExpression,
		TokenTypeNumber:    p.parseNumericLiteralExpression,
		TokenTypeFalse:     p.parseBooleanLiteralExpression,
		TokenTypeNot:       p.parseUnary(expressions.NewNot),
		TokenTypeNull:      p.parseNullLiteralExpression,
		TokenTypeTrue:      p.parseBooleanLiteralExpression,
		TokenTypeMinus:     p.parseUnary(expressions.NewUnaryMinus),
		TokenTypeLeftParen: p.parseParenthesizedExpression,
		TokenTypePlus:      p.parseUnary(expressions.NewUnaryPlus),
	}

	p.infixParsers = map[TokenType]infixParserFunc{
		TokenTypeAnd: p.parseBinary(PrecedenceConditionalAnd, expressions.NewAnd),
		// TokenTypeIs: nil, // TODO - also not, null, isnull, notnull
		TokenTypeOr:    p.parseBinary(PrecedenceConditionalOr, expressions.NewOr),
		TokenTypeMinus: p.parseBinary(PrecedenceAdditive, expressions.NewSubtraction),
		// TokenTypeDot:         nil, // TODO - make projection more general
		// TokenTypeLeftParen:   nil, // TODO - function calls
		TokenTypeAsterisk:           p.parseBinary(PrecedenceMultiplicative, expressions.NewMultiplication),
		TokenTypeSlash:              p.parseBinary(PrecedenceMultiplicative, expressions.NewDivision),
		TokenTypePlus:               p.parseBinary(PrecedenceAdditive, expressions.NewAddition),
		TokenTypeLessThan:           p.parseBinary(PrecedenceComparison, expressions.NewLessThan),
		TokenTypeEquals:             p.parseBinary(PrecedenceEquality, expressions.NewEquals),
		TokenTypeGreaterThan:        p.parseBinary(PrecedenceComparison, expressions.NewGreaterThan),
		TokenTypeLessThanOrEqual:    p.parseBinary(PrecedenceComparison, expressions.NewGreaterThanEquals),
		TokenTypeNotEquals:          negate(p.parseBinary(PrecedenceEquality, expressions.NewEquals)),
		TokenTypeGreaterThanOrEqual: p.parseBinary(PrecedenceComparison, expressions.NewGreaterThanEquals),
	}
}

func (p *parser) parseStatement() (relations.Relation, error) {
	if p.advanceIf(isType(TokenTypeSelect)) {
		return p.parseSelect()
	}

	return nil, fmt.Errorf("expected start of statement")
}

func (p *parser) parseSelect() (relations.Relation, error) {
	selectExpressions, err := p.parseSelectExpressions()
	if err != nil {
		return nil, err
	}

	relation, err := p.parseFrom()
	if err != nil {
		return nil, err
	}

	whereExpression, hasWhere, err := p.parseWhereClause()
	if err != nil {
		return nil, err
	}
	orderExpression, hasOrder, err := p.parseOrderByClause()
	if err != nil {
		return nil, err
	}
	limitValue, hasLimit, err := p.parseLimitClause()
	if err != nil {
		return nil, err
	}
	offsetValue, hasOffset, err := p.parseOffsetClause()
	if err != nil {
		return nil, err
	}

	if hasWhere {
		relation = relations.NewFilter(relation, whereExpression)
	}
	if hasOrder {
		relation = relations.NewOrder(relation, orderExpression)
	}
	if hasOffset {
		relation = relations.NewOffset(relation, offsetValue)
	}
	if hasLimit {
		relation = relations.NewLimit(relation, limitValue)
	}
	if len(selectExpressions) > 0 {
		relation = relations.NewProjection(relation, selectExpressions)
	}

	return relation, nil
}

func (p *parser) parseSelectExpressions() (aliasedExpressions []relations.AliasedExpression, _ error) {
	if p.advanceIf(isType(TokenTypeAsterisk)) {
		return nil, nil
	}

	for {
		expression, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		alias, err := p.parseAlias(expression)
		if err != nil {
			return nil, err
		}

		aliasedExpressions = append(aliasedExpressions, relations.AliasedExpression{
			Alias:      alias,
			Expression: expression,
		})

		if !p.advanceIf(isType(TokenTypeComma)) {
			break
		}
	}

	return aliasedExpressions, nil
}

type named interface {
	Name() string
}

func (p *parser) parseAlias(expression expressions.Expression) (string, error) {
	alias := "?column?"
	if named, ok := expression.(named); ok {
		alias = named.Name()
	}

	if p.advanceIf(isType(TokenTypeAs)) {
		aliasToken, err := p.mustAdvance(isType(TokenTypeIdent))
		if err != nil {
			return "", err
		}

		alias = aliasToken.Text
	}

	return alias, nil
}

func (p *parser) parseFrom() (relation relations.Relation, _ error) {
	if _, err := p.mustAdvance(isType(TokenTypeFrom)); err != nil {
		return nil, err
	}

	for {
		fromToken, err := p.mustAdvance(isType(TokenTypeIdent))
		if err != nil {
			return nil, err
		}
		left, ok := p.builtins[fromToken.Text]
		if !ok {
			return nil, fmt.Errorf("unknown table %s", fromToken.Text)
		}
		if p.current().Type == TokenTypeIdent {
			left = relations.NewAlias(left, p.advance().Text)
		}

		if relation == nil {
			relation = left
		} else {
			relation = relations.NewJoin(relation, left, nil)
		}

		for p.advanceIf(isType(TokenTypeJoin)) {
			fromToken, err := p.mustAdvance(isType(TokenTypeIdent))
			if err != nil {
				return nil, err
			}
			right, ok := p.builtins[fromToken.Text]
			if !ok {
				return nil, fmt.Errorf("unknown table %s", fromToken.Text)
			}
			if p.current().Type == TokenTypeIdent {
				right = relations.NewAlias(right, p.advance().Text)
			}

			var condition expressions.BoolExpression
			if p.advanceIf(isType(TokenTypeOn)) {
				rawCondition, err := p.parseExpression(0)
				if err != nil {
					return nil, err
				}

				condition = expressions.Bool(rawCondition)
			}

			relation = relations.NewJoin(relation, right, condition)
		}

		if !p.advanceIf(isType(TokenTypeComma)) {
			break
		}
	}

	return relation, nil
}

func (p *parser) parseWhereClause() (expressions.BoolExpression, bool, error) {
	if !p.advanceIf(isType(TokenTypeWhere)) {
		return nil, false, nil
	}

	whereExpression, err := p.parseExpression(0)
	if err != nil {
		return nil, false, err
	}

	return expressions.Bool(whereExpression), true, nil
}

func (p *parser) parseOrderByClause() (expressions.Expression, bool, error) {
	if !p.advanceIf(isType(TokenTypeOrder)) {
		return nil, false, nil
	}

	if _, err := p.mustAdvance(isType(TokenTypeBy)); err != nil {
		return nil, false, err
	}

	orderExpression, err := p.parseExpression(0)
	if err != nil {
		return nil, false, err
	}

	return orderExpression, true, nil
}

func (p *parser) parseLimitClause() (int, bool, error) {
	if !p.advanceIf(isType(TokenTypeLimit)) {
		return 0, false, nil
	}

	limitToken, err := p.mustAdvance(isType(TokenTypeNumber))
	if err != nil {
		return 0, false, err
	}

	limitValue, err := strconv.Atoi(limitToken.Text)
	return limitValue, true, err
}

func (p *parser) parseOffsetClause() (int, bool, error) {
	if !p.advanceIf(isType(TokenTypeOffset)) {
		return 0, false, nil
	}

	limitToken, err := p.mustAdvance(isType(TokenTypeNumber))
	if err != nil {
		return 0, false, err
	}

	limitValue, err := strconv.Atoi(limitToken.Text)
	return limitValue, true, err
}

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
		return nil, fmt.Errorf("expected expression")
	}

	return parseFunc(current)
}

func (p *parser) parseExpressionSuffix(expression expressions.Expression, precedence Precedence) (_ expressions.Expression, err error) {
	for {
		tokenType := p.current().Type
		tokenPrecedence, ok := precedenceMap[tokenType]
		if !ok {
			break
		}
		if tokenPrecedence < precedence {
			break
		}
		parseFunc, ok := p.infixParsers[tokenType]
		if !ok {
			break
		}

		expression, err = parseFunc(expression, p.advance())
		if err != nil {
			return nil, err
		}
	}

	return expression, nil
}

//
// Prefix parsers

func (p *parser) parseNamedExpression(token Token) (expressions.Expression, error) {
	if !p.advanceIf(isType(TokenTypeDot)) {
		return expressions.NewNamed("", token.Text), nil
	}

	qualifiedNameToken, err := p.mustAdvance(isType(TokenTypeIdent))
	if err != nil {
		return nil, err
	}

	return expressions.NewNamed(token.Text, qualifiedNameToken.Text), nil
}

func (p *parser) parseNumericLiteralExpression(token Token) (expressions.Expression, error) {
	value, err := strconv.Atoi(token.Text)
	if err != nil {
		return nil, err
	}

	return expressions.NewConstant(value), nil
}

func (p *parser) parseBooleanLiteralExpression(token Token) (expressions.Expression, error) {
	return expressions.NewConstant(token.Type == TokenTypeTrue), nil
}

func (p *parser) parseNullLiteralExpression(token Token) (expressions.Expression, error) {
	return expressions.NewConstant(nil), nil
}

func (p *parser) parseParenthesizedExpression(token Token) (expressions.Expression, error) {
	inner, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	if _, err := p.mustAdvance(isType(TokenTypeRightParen)); err != nil {
		return nil, err
	}

	return inner, nil
}

func (p *parser) parseUnary(factory unaryExpressionParserFunc) prefixParserFunc {
	return func(token Token) (expressions.Expression, error) {
		expression, err := p.parseExpression(PrecedenceUnary)
		if err != nil {
			return nil, err
		}

		return factory(expression), nil
	}
}

//
// Infix parsers

func (p *parser) parseBinary(precedence Precedence, factory binaryExpressionParserFunc) infixParserFunc {
	return func(left expressions.Expression, token Token) (expressions.Expression, error) {
		right, err := p.parseExpression(precedence)
		if err != nil {
			return nil, err
		}

		return factory(left, right), nil
	}
}

func (p *parser) current() Token {
	if p.cursor >= len(p.tokens) {
		return InvalidToken
	}

	return p.tokens[p.cursor]
}

func (p *parser) advance() Token {
	r := p.current()
	p.cursor++
	return r
}

func (p *parser) advanceIf(filter tokenFilterFunc) bool {
	if !filter(p.current()) {
		return false
	}

	p.cursor++
	return true
}

func (p *parser) mustAdvance(filter tokenFilterFunc) (Token, error) {
	current := p.advance()
	if !filter(current) {
		return InvalidToken, fmt.Errorf("unexpected token")
	}

	return current, nil
}

func isType(tokenType TokenType) tokenFilterFunc {
	return func(t Token) bool {
		return t.Type == tokenType
	}
}

func negate(parserFunc infixParserFunc) infixParserFunc {
	return func(left expressions.Expression, token Token) (expressions.Expression, error) {
		expression, err := parserFunc(left, token)
		if err != nil {
			return nil, err
		}

		return expressions.NewNot(expression), nil
	}
}
