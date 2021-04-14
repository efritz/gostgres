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
		fromToken, err := p.mustAdvance(func(t Token) bool { return t.Type == TokenTypeIdent })
		if err != nil {
			return nil, err
		}

		r, ok := builtins[fromToken.Text]
		if !ok {
			return nil, fmt.Errorf("unknown table %s", fromToken.Text)
		}

		if relation == nil {
			relation = r
		} else {
			// TODO - parse join condition
			relation = relations.NewJoin(relation, r, nil)
		}

		if !p.advanceIf(isType(TokenTypeComma)) {
			break
		}
	}

	// TODO - support WHERE
	// TODO - support ORDER BY
	// TODO - flip order of application (but not parsing) of limit/offset

	if p.advanceIf(isKeyword("LIMIT")) {
		limitToken, err := p.mustAdvance(isType(TokenTypeNumber))
		if err != nil {
			return nil, err
		}

		limitValue, _ := strconv.Atoi(limitToken.Text)
		relation = relations.NewLimit(relation, limitValue)
	}

	if p.advanceIf(isKeyword("OFFSET")) {
		offsetToken, err := p.mustAdvance(isType(TokenTypeNumber))
		if err != nil {
			return nil, err
		}

		offsetValue, _ := strconv.Atoi(offsetToken.Text)
		relation = relations.NewOffset(relation, offsetValue)
	}

	if len(aliasedExpressions) > 0 {
		relation = relations.NewProjection(relation, aliasedExpressions)
	}

	return relation, nil
}

func (p *parser) parseAliasedExpressions() ([]relations.AliasedExpression, error) {
	var aliasedExpressions []relations.AliasedExpression
	for {
		expression, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		// TODO - should use column name when obvious
		alias := "?column?"

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

func (p *parser) parseExpression() (expressions.Expression, error) {
	current := p.advance()

	// TODO - load matchers into a map
	// TODO - implement precedence parsing

	if current.Type == TokenTypeIdent {
		return p.parseNamed(current)
	}
	if current.Type == TokenTypeNumber {
		return p.parseConstantNumber(current)
	}

	return nil, fmt.Errorf("expected expression")
}

func (p *parser) parseNamed(nameToken Token) (expressions.Expression, error) {
	if !p.advanceIf(isType(TokenTypeDot)) {
		return expressions.NewNamed("", nameToken.Text), nil
	}

	secondNameToken, err := p.mustAdvance(isType(TokenTypeIdent))
	if err != nil {
		return nil, err
	}

	return expressions.NewNamed(nameToken.Text, secondNameToken.Text), nil
}

func (p *parser) parseConstantNumber(numberToken Token) (expressions.Expression, error) {
	value, err := strconv.Atoi(numberToken.Text)
	if err != nil {
		return nil, err
	}

	return expressions.NewConstant(value), nil
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
