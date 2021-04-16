package syntax

import (
	"fmt"
	"strconv"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/relations"
	"github.com/efritz/gostgres/internal/shared"
)

func Parse(tokens []Token, tables map[string]*relations.Table) (relations.Relation, error) {
	parser := &parser{
		tokens: tokens,
		tables: tables,
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
	tables        map[string]*relations.Table
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
		TokenTypeString:    p.parseStringLiteralExpression,
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
		TokenTypeLessThanOrEqual:    p.parseBinary(PrecedenceComparison, expressions.NewLessThanEquals),
		TokenTypeNotEquals:          negate(p.parseBinary(PrecedenceEquality, expressions.NewEquals)),
		TokenTypeGreaterThanOrEqual: p.parseBinary(PrecedenceComparison, expressions.NewGreaterThanEquals),
	}
}

// consumes: `SELECT` select
// consumes: `INSERT` `INTO` ident select
// consumes: `INSERT` `INTO` ident `VALUES` `(` expr [, ...] `)` [, ...]
func (p *parser) parseStatement() (relations.Relation, error) {
	if p.advanceIf(isType(TokenTypeSelect)) {
		return p.parseSelect()
	}

	if p.advanceIf(isType(TokenTypeInsert)) {
		return p.parseInsert()
	}

	return nil, fmt.Errorf("expected start of statement (near %s)", p.current().Text)
}

//
// Select expressions

// consumes: select_expressions from [where] [order] [limit] [offset]
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
		relation, err = relations.NewProjection(relation, selectExpressions)
		if err != nil {
			return nil, err
		}
	}

	return relation, nil
}

// consumes: `*`
// consumes: alias_expression [, ...]
func (p *parser) parseSelectExpressions() (aliasedExpressions []relations.ProjectionExpression, _ error) {
	if p.advanceIf(isType(TokenTypeAsterisk)) {
		return nil, nil
	}

	for {
		aliasedExpression, err := p.parseSelectExpression()
		if err != nil {
			return nil, err
		}

		aliasedExpressions = append(aliasedExpressions, aliasedExpression)

		if !p.advanceIf(isType(TokenTypeComma)) {
			break
		}
	}

	return aliasedExpressions, nil
}

// consumes: ident `.` `*`
// consumes: expression [alias]
func (p *parser) parseSelectExpression() (relations.ProjectionExpression, error) {
	nameToken := p.current()
	if p.advanceIf(isType(TokenTypeIdent), isType(TokenTypeDot), isType(TokenTypeAsterisk)) {
		return relations.NewWildcardProjectionExpression(nameToken.Text), nil
	}

	expression, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	alias, err := p.parseColumnAlias(expression)
	if err != nil {
		return nil, err
	}

	return relations.NewAliasProjectionExpression(expression, alias), nil
}

// consumes: nothing
// consumes: ident
// consumes: `AS` ident
func (p *parser) parseColumnAlias(expression expressions.Expression) (string, error) {
	type named interface {
		Name() string
	}

	alias := "?column?"
	if named, ok := expression.(named); ok {
		alias = named.Name()
	}

	if p.advanceIf(isType(TokenTypeAs)) {
		aliasToken, err := p.mustAdvance(isType(TokenTypeIdent))
		if err != nil {
			return "", err
		}

		return aliasToken.Text, nil
	}

	nameToken := p.current()
	if p.advanceIf(isType(TokenTypeIdent)) {
		return nameToken.Text, nil
	}

	return alias, nil
}

// consumes: `FROM` table_expression [, ...]
func (p *parser) parseFrom() (relation relations.Relation, _ error) {
	if _, err := p.mustAdvance(isType(TokenTypeFrom)); err != nil {
		return nil, err
	}

	for {
		tableExpression, err := p.parseTableExpression()
		if err != nil {
			return nil, err
		}

		if relation == nil {
			relation = tableExpression
		} else {
			relation = relations.NewJoin(relation, tableExpression, nil)
		}

		if !p.advanceIf(isType(TokenTypeComma)) {
			break
		}
	}

	return relation, nil
}

// consumes: base_table_expression [`JOIN` join [`JOIN` ...]]
func (p *parser) parseTableExpression() (relations.Relation, error) {
	relation, err := p.parseBaseTableExpression()
	if err != nil {
		return nil, err
	}

	for p.advanceIf(isType(TokenTypeJoin)) {
		relation, err = p.parseJoin(relation)
		if err != nil {
			return nil, err
		}
	}

	return relation, nil
}

// consumes: ident [alias]
// consumes: (select) alias
// consumes: (`VALUES` `(` expr [, ...] `)` [, ...]) alias
// consumes: (table_expression) [alias]
func (p *parser) parseBaseTableExpression() (relations.Relation, error) {
	expectParen := false
	requireAlias := false
	parseFunc := p.parseTableReference

	if p.advanceIf(isType(TokenTypeLeftParen)) {
		if p.current().Type == TokenTypeSelect || p.current().Type == TokenTypeValues {
			requireAlias = true
			parseFunc = p.parseSelectOrValues
		} else {
			parseFunc = p.parseTableExpression
		}

		expectParen = true
	}

	relation, err := parseFunc()
	if err != nil {
		return nil, err
	}

	if expectParen {
		if _, err := p.mustAdvance(isType(TokenTypeRightParen)); err != nil {
			return nil, err
		}
	}

	if p.advanceIf(isType(TokenTypeAs)) {
		if p.current().Type != TokenTypeIdent {
			return nil, fmt.Errorf("expected alias (near %s)", p.current().Text)
		}
	}

	if p.current().Type == TokenTypeIdent {
		relation = relations.NewAlias(relation, p.advance().Text)
	} else if requireAlias {
		return nil, fmt.Errorf("expected subselect alias (near %s)", p.current().Text)
	}

	return relation, nil
}

// consumes: ident
func (p *parser) parseTableReference() (relations.Relation, error) {
	nameToken, err := p.mustAdvance(isType(TokenTypeIdent))
	if err != nil {
		return nil, err
	}

	table, ok := p.tables[nameToken.Text]
	if !ok {
		return nil, fmt.Errorf("unknown table %s", nameToken.Text)
	}

	return relations.NewData(nameToken.Text, table), nil
}

// consumes: table_expression [`ON` expression]
func (p *parser) parseJoin(relation relations.Relation) (relations.Relation, error) {
	right, err := p.parseTableExpression()
	if err != nil {
		return nil, err
	}

	var condition expressions.Expression
	if p.advanceIf(isType(TokenTypeOn)) {
		rawCondition, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		condition = rawCondition
	}

	return relations.NewJoin(relation, right, condition), nil
}

// consumes: [`WHERE` expression]
func (p *parser) parseWhereClause() (expressions.Expression, bool, error) {
	if !p.advanceIf(isType(TokenTypeWhere)) {
		return nil, false, nil
	}

	whereExpression, err := p.parseExpression(0)
	if err != nil {
		return nil, false, err
	}

	return whereExpression, true, nil
}

// consumes: [`ORDER` `BY` expression]
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

// consumes: [`LIMIT` expression]
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

// consumes: [`OFFSET` expression]
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

//
// Insert statements

// consumes: `INSERT` `INTO` alias? ident { `(` ident [, ...]`)` }? select
// consumes: `INSERT` `INTO` alias? ident { `(` ident [, ...]`)` }? `VALUES` `(` expr [, ...] `)` [, ...]
func (p *parser) parseInsert() (relations.Relation, error) {
	if _, err := p.mustAdvance(isType(TokenTypeInto)); err != nil {
		return nil, err
	}

	nameToken, err := p.mustAdvance(isType(TokenTypeIdent))
	if err != nil {
		return nil, err
	}

	table, ok := p.tables[nameToken.Text]
	if !ok {
		return nil, fmt.Errorf("unknown table %s", nameToken.Text)
	}

	if p.advanceIf(isType(TokenTypeAs)) {
		if p.current().Type != TokenTypeIdent {
			return nil, fmt.Errorf("expected alias (near %s)", p.current().Text)
		}
	}

	alias := ""
	aliasToken := p.current()
	if p.advanceIf(isType(TokenTypeIdent)) {
		alias = aliasToken.Text
	}

	var columnNames []string
	if p.current().Type == TokenTypeLeftParen && p.peek(1).Type == TokenTypeIdent {
		p.advance()

		for {
			nameToken, err := p.mustAdvance(isType(TokenTypeIdent))
			if err != nil {
				return nil, err
			}

			columnNames = append(columnNames, nameToken.Text)

			if !p.advanceIf(isType(TokenTypeComma)) {
				break
			}
		}

		if _, err := p.mustAdvance(isType(TokenTypeRightParen)); err != nil {
			return nil, err
		}
	}

	relation, err := p.parseSelectOrValues()
	if err != nil {
		return nil, err
	}

	var returningExpressions []relations.ProjectionExpression
	if p.advanceIf(isType(TokenTypeReturning)) {
		returningExpressions, err = p.parseSelectExpressions()
		if err != nil {
			return nil, err
		}

		if returningExpressions == nil {
			returningExpressions = []relations.ProjectionExpression{
				relations.NewWildcardProjectionExpression(nameToken.Text),
			}
		}
	}

	return relations.NewInsert(relation, table, nameToken.Text, alias, columnNames, returningExpressions)
}

// consumes: `SELECT` select
// consumes: `VALUES` `(` expr [, ...] `)` [, ...]
func (p *parser) parseSelectOrValues() (relations.Relation, error) {
	if p.advanceIf(isType(TokenTypeSelect)) {
		return p.parseSelect()
	}

	return p.parseValues()
}

// consumes: `VALUES` `(` expr [, ...] `)` [, ...]
func (p *parser) parseValues() (relations.Relation, error) {
	if _, err := p.mustAdvance(isType(TokenTypeValues)); err != nil {
		return nil, err
	}

	return p.parseValuesList()
}

// consumes: (` expr [, ...] `)` [, ...]
func (p *parser) parseValuesList() (relations.Relation, error) {
	var allRows [][]interface{}
	for {
		if _, err := p.mustAdvance(isType(TokenTypeLeftParen)); err != nil {
			return nil, err
		}

		var values []interface{}
		for {
			expression, err := p.parseExpression(PrecedenceAny)
			if err != nil {
				return nil, err
			}

			value, err := expression.ValueFrom(shared.Row{})
			if err != nil {

				return nil, err
			}

			values = append(values, value)

			if !p.advanceIf(isType(TokenTypeComma)) {
				break
			}
		}

		if _, err := p.mustAdvance(isType(TokenTypeRightParen)); err != nil {
			return nil, err
		}

		allRows = append(allRows, values)

		if !p.advanceIf(isType(TokenTypeComma)) {
			break
		}
	}

	fields := make([]shared.Field, 0, len(allRows[0]))
	for i := range allRows[0] {
		fields = append(fields, shared.Field{
			RelationName: "",
			Name:         fmt.Sprintf("column%d", i+1),
		})
	}

	rows := shared.NewRows(fields)

	for _, values := range allRows {
		var err error
		rows, err = rows.AddValues(values)
		if err != nil {
			return nil, err
		}
	}

	return relations.NewData("", relations.NewTable(rows)), nil
}

//
// Value expressions

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

// consumes: [`.` ident]
func (p *parser) parseNamedExpression(token Token) (expressions.Expression, error) {
	if !p.advanceIf(isType(TokenTypeDot)) {
		return expressions.NewNamed(shared.Field{
			Name: token.Text,
		}), nil
	}

	qualifiedNameToken, err := p.mustAdvance(isType(TokenTypeIdent))
	if err != nil {
		return nil, err
	}

	return expressions.NewNamed(shared.Field{
		RelationName: token.Text,
		Name:         qualifiedNameToken.Text,
	}), nil
}

// consumes: nothing
func (p *parser) parseNumericLiteralExpression(token Token) (expressions.Expression, error) {
	value, err := strconv.Atoi(token.Text)
	if err != nil {
		return nil, err
	}

	return expressions.NewConstant(value), nil
}

// consumes: nothing
func (p *parser) parseStringLiteralExpression(token Token) (expressions.Expression, error) {
	return expressions.NewConstant(token.Text), nil
}

// consumes: nothing
func (p *parser) parseBooleanLiteralExpression(token Token) (expressions.Expression, error) {
	return expressions.NewConstant(token.Type == TokenTypeTrue), nil
}

// consumes: nothing
func (p *parser) parseNullLiteralExpression(token Token) (expressions.Expression, error) {
	return expressions.NewConstant(nil), nil
}

// consumes: expression `)`
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
	return p.peek(0)
}

func (p *parser) peek(n int) Token {
	if p.cursor+n >= len(p.tokens) {
		return InvalidToken
	}

	return p.tokens[p.cursor+n]
}

func (p *parser) advance() Token {
	r := p.current()
	p.cursor++
	return r
}

func (p *parser) advanceIf(filters ...tokenFilterFunc) bool {
	start := p.cursor
	for _, filter := range filters {
		if !filter(p.current()) {
			p.cursor = start
			return false
		}

		p.cursor++
	}

	return true
}

func (p *parser) mustAdvance(filter tokenFilterFunc) (Token, error) {
	current := p.advance()
	if !filter(current) {
		return InvalidToken, fmt.Errorf("unexpected token (near %s)", current.Text)
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
