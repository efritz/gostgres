package syntax

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/relations"
)

func Parse(tokens []Token, builtins map[string]relations.Relation) (relations.Relation, error) {
	parser := &parser{tokens: tokens}
	return parser.parseRelation(builtins)
}

type parser struct {
	tokens []Token
	cursor int
}

func (p *parser) parseRelation(builtins map[string]relations.Relation) (relations.Relation, error) {
	if !p.advanceIf(isKeyword("SELECT")) {
		return nil, fmt.Errorf("expected start of statement")
	}

	var aliasedExpressions []relations.AliasedExpression
	if !p.advanceIf(isType(TokenTypeAsterisk)) {
		var err error
		aliasedExpressions, err = p.parseAliasedExpressions()
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.mustAdvance(isKeyword("FROM")); err != nil {
		return nil, err
	}

	var relation relations.Relation
	for {
		fromToken, err := p.mustAdvance(isType(TokenTypeIdent))
		if err != nil {
			return nil, err
		}
		left, ok := builtins[fromToken.Text]
		if !ok {
			return nil, fmt.Errorf("unknown table %s", fromToken.Text)
		}

		if relation == nil {
			relation = left
		} else {
			relation = relations.NewJoin(relation, left, nil)
		}

		if p.advanceIf(isKeyword("JOIN")) {
			fromToken, err := p.mustAdvance(isType(TokenTypeIdent))
			if err != nil {
				return nil, err
			}
			right, ok := builtins[fromToken.Text]
			if !ok {
				return nil, fmt.Errorf("unknown table %s", fromToken.Text)
			}

			var condition expressions.BoolExpression
			if p.advanceIf(isKeyword("ON")) {
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

	if p.advanceIf(isKeyword("WHERE")) {
		condition, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		relation = relations.NewFilter(relation, expressions.Bool(condition))
	}

	if p.advanceIf(isKeyword("ORDER")) {
		if _, err := p.mustAdvance(isKeyword("BY")); err != nil {
			return nil, err
		}

		orderExpression, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		relation = relations.NewOrder(relation, expressions.Int(orderExpression))
	}

	hasLimit := false
	limitValue := 0

	if p.advanceIf(isKeyword("LIMIT")) {
		limitToken, err := p.mustAdvance(isType(TokenTypeNumber))
		if err != nil {
			return nil, err
		}

		// save and apply after offset
		hasLimit = true
		limitValue, _ = strconv.Atoi(limitToken.Text)
	}

	if p.advanceIf(isKeyword("OFFSET")) {
		offsetToken, err := p.mustAdvance(isType(TokenTypeNumber))
		if err != nil {
			return nil, err
		}

		offsetValue, _ := strconv.Atoi(offsetToken.Text)
		relation = relations.NewOffset(relation, offsetValue)
	}

	if hasLimit {
		relation = relations.NewLimit(relation, limitValue)
	}

	if len(aliasedExpressions) > 0 {
		relation = relations.NewProjection(relation, aliasedExpressions)
	}

	return relation, nil
}

type named interface {
	Name() string
}

func (p *parser) parseAliasedExpressions() ([]relations.AliasedExpression, error) {
	var aliasedExpressions []relations.AliasedExpression
	for {
		expression, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		alias := "?column?"
		if named, ok := expression.(named); ok {
			alias = named.Name()
		}

		if p.advanceIf(isKeyword("AS")) {
			aliasToken, err := p.mustAdvance(isType(TokenTypeIdent))
			if err != nil {
				return nil, err
			}

			alias = aliasToken.Text
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

type prefixParser interface {
	Parse(token Token) (expressions.Expression, error)
}

type prefixParserFunc func(token Token) (expressions.Expression, error)

func (f prefixParserFunc) Parse(token Token) (expressions.Expression, error) {
	return f(token)
}

type infixParser interface {
	Precedence() int
	Parse(left expressions.Expression, token Token) (expressions.Expression, error)
}

type genericInfixParser struct {
	precedence int
	f          func(left expressions.Expression, token Token) (expressions.Expression, error)
}

func (p *genericInfixParser) Precedence() int {
	return p.precedence
}

func (p *genericInfixParser) Parse(left expressions.Expression, token Token) (expressions.Expression, error) {
	return p.f(left, token)
}

func (p *parser) parseExpression(precedence int) (expressions.Expression, error) {
	prefixParsers := map[TokenType]prefixParser{
		TokenTypeLeftParen: prefixParserFunc(p.parseParenthesizedExpression),
		TokenTypeIdent:     prefixParserFunc(p.parseNamedExpression),
		TokenTypeNumber:    prefixParserFunc(p.parseNumericLiteralExpression),
	}

	infixParsers := map[TokenType]infixParser{
		TokenTypeEquals: &genericInfixParser{1, p.parseEqualsExpression},
		TokenTypePlus:   &genericInfixParser{3, p.parsePlusExpression},
	}

	current := p.advance()
	parser, ok := prefixParsers[current.Type]
	if !ok {
		return nil, fmt.Errorf("expected expression")
	}
	left, err := parser.Parse(current)
	if err != nil {
		return nil, err
	}

	for {
		parser, ok := infixParsers[p.current().Type]
		if !ok || precedence >= parser.Precedence() {
			break
		}

		left, err = parser.Parse(left, p.advance())
		if err != nil {
			return nil, err
		}
	}

	return left, nil
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

func (p *parser) parseEqualsExpression(left expressions.Expression, token Token) (expressions.Expression, error) {
	right, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	return expressions.NewEquals(left, right), nil
}

func (p *parser) parsePlusExpression(left expressions.Expression, token Token) (expressions.Expression, error) {
	right, err := p.parseExpression(3)
	if err != nil {
		return nil, err
	}

	return expressions.NewSum(left, right), nil
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

func (p *parser) advanceIf(filter func(t Token) bool) bool {
	if !filter(p.current()) {
		return false
	}

	p.cursor++
	return true
}

func (p *parser) mustAdvance(filter func(t Token) bool) (Token, error) {
	current := p.advance()
	if !filter(current) {
		return InvalidToken, fmt.Errorf("unexpected token")
	}

	return current, nil
}

func isType(tokenType TokenType) func(t Token) bool {
	return func(t Token) bool {
		return t.Type == tokenType
	}
}

func isKeyword(text string) func(t Token) bool {
	return func(t Token) bool {
		return t.Type == TokenTypeKeyword && strings.ToUpper(t.Text) == text
	}
}
