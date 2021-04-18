package syntax

import (
	"fmt"
	"strconv"

	"github.com/efritz/gostgres/internal/expressions"
	nodes "github.com/efritz/gostgres/internal/nodes"
	"github.com/efritz/gostgres/internal/shared"
)

func Parse(tokens []Token, tables map[string]*nodes.Table) (nodes.Node, error) {
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
	tokens           []Token
	cursor           int
	tables           map[string]*nodes.Table
	statementParsers map[TokenType]statementParserFunc
	prefixParsers    map[TokenType]prefixParserFunc
	infixParsers     map[TokenType]infixParserFunc
}

type tokenFilterFunc func(token Token) bool
type statementParserFunc func(token Token) (nodes.Node, error)
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
	p.statementParsers = map[TokenType]statementParserFunc{
		TokenTypeSelect: p.parseSelect,
		TokenTypeInsert: p.parseInsert,
		TokenTypeUpdate: p.parseUpdate,
		TokenTypeDelete: p.parseDelete,
	}

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
		TokenTypeAnd:                p.parseBinary(PrecedenceConditionalAnd, expressions.NewAnd),
		TokenTypeOr:                 p.parseBinary(PrecedenceConditionalOr, expressions.NewOr),
		TokenTypeMinus:              p.parseBinary(PrecedenceAdditive, expressions.NewSubtraction),
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

// statement := `SELECT` select
//            | `INSERT` insert
//            | `UPDATE` update
//            | `DELETE` delete
func (p *parser) parseStatement() (nodes.Node, error) {
	token := p.current()
	for tokenType, parser := range p.statementParsers {
		if p.advanceIf(isType(tokenType)) {
			return parser(token)
		}
	}

	return nil, fmt.Errorf("expected start of statement (near %s)", p.current().Text)
}

//
// Select expressions

// select := selectExpressions from where order limit offset
func (p *parser) parseSelect(token Token) (nodes.Node, error) {
	selectExpressions, err := p.parseSelectExpressions()
	if err != nil {
		return nil, err
	}

	node, err := p.parseFrom()
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
		node = nodes.NewFilter(node, whereExpression)
	}
	if hasOrder {
		node = nodes.NewOrder(node, orderExpression)
	}
	if hasOffset {
		node = nodes.NewOffset(node, offsetValue)
	}
	if hasLimit {
		node = nodes.NewLimit(node, limitValue)
	}

	node, err = nodes.NewProjection(node, selectExpressions)
	if err != nil {
		return nil, err
	}

	return node, nil
}

// selectExpressions := `*`
//                    | selectExpression [, ...]
func (p *parser) parseSelectExpressions() (aliasedExpressions []nodes.ProjectionExpression, _ error) {
	if p.advanceIf(isType(TokenTypeAsterisk)) {
		return []nodes.ProjectionExpression{nodes.NewWildcardProjectionExpression()}, nil
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

// selectExpression := ident `.` `*`
//                   | expression columnAlias
func (p *parser) parseSelectExpression() (nodes.ProjectionExpression, error) {
	nameToken := p.current()
	if p.advanceIf(isType(TokenTypeIdent), isType(TokenTypeDot), isType(TokenTypeAsterisk)) {
		return nodes.NewTableWildcardProjectionExpression(nameToken.Text), nil
	}

	expression, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	alias, err := p.parseColumnAlias(expression)
	if err != nil {
		return nil, err
	}

	return nodes.NewAliasProjectionExpression(expression, alias), nil
}

// columnAlias := nothing
//              | [`AS`] ident
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

// from := `FROM` ( tableExpression [, ...] )
func (p *parser) parseFrom() (node nodes.Node, _ error) {
	if _, err := p.mustAdvance(isType(TokenTypeFrom)); err != nil {
		return nil, err
	}

	for {
		tableExpression, err := p.parseTableExpression()
		if err != nil {
			return nil, err
		}

		if node == nil {
			node = tableExpression
		} else {
			node = nodes.NewJoin(node, tableExpression, nil)
		}

		if !p.advanceIf(isType(TokenTypeComma)) {
			break
		}
	}

	return node, nil
}

// tableExpression := baseTableExpression [( `JOIN` join [...] )]
func (p *parser) parseTableExpression() (nodes.Node, error) {
	node, err := p.parseBaseTableExpression()
	if err != nil {
		return nil, err
	}

	for p.advanceIf(isType(TokenTypeJoin)) {
		node, err = p.parseJoin(node)
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

// baseTableExpression := tableReference [[`AS`] ident]
//                      | `(` selectOrValues `)` [`AS`] ident
//                      | `(` tableExpression `)` [[`AS`] ident]
func (p *parser) parseBaseTableExpression() (nodes.Node, error) {
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

	node, err := parseFunc()
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
		node = nodes.NewAlias(node, p.advance().Text)
	} else if requireAlias {
		return nil, fmt.Errorf("expected subselect alias (near %s)", p.current().Text)
	}

	return node, nil
}

// tableReference := ident
func (p *parser) parseTableReference() (nodes.Node, error) {
	nameToken, err := p.mustAdvance(isType(TokenTypeIdent))
	if err != nil {
		return nil, err
	}

	table, ok := p.tables[nameToken.Text]
	if !ok {
		return nil, fmt.Errorf("unknown table %s", nameToken.Text)
	}

	return nodes.NewData(table), nil
}

// join := tableExpression [`ON` expression]
func (p *parser) parseJoin(node nodes.Node) (nodes.Node, error) {
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

	return nodes.NewJoin(node, right, condition), nil
}

// where := nothing
//        | `WHERE` expression
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

// where := nothing
//        | `ORDER` `BY` ( expression [( `ASC` | `DESC` )] [, ...] )
func (p *parser) parseOrderByClause() (nodes.OrderExpression, bool, error) {
	if !p.advanceIf(isType(TokenTypeOrder)) {
		return nil, false, nil
	}

	if _, err := p.mustAdvance(isType(TokenTypeBy)); err != nil {
		return nil, false, err
	}

	var expressions []nodes.FieldExpression
	for {
		orderExpression, err := p.parseExpression(0)
		if err != nil {
			return nil, false, err
		}

		reverse := false
		if !p.advanceIf(isType(TokenTypeAscending)) {
			if p.advanceIf(isType(TokenTypeDescending)) {
				reverse = true
			}
		}

		expressions = append(expressions, nodes.FieldExpression{
			Expression: orderExpression,
			Reverse:    reverse,
		})

		if !p.advanceIf(isType(TokenTypeComma)) {
			break
		}
	}

	return nodes.NewOrderExpression(expressions), true, nil
}

// limit := nothing
//        | `LIMIT` expression
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

// offset := nothing
//         | `LIMIT` expression
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

// insert := `INTO` ident [[`AS` ident]] [`(` ident [, ...] `)`] selectOrValues [`RETURNING` selectExpressions]
func (p *parser) parseInsert(token Token) (nodes.Node, error) {
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

	node, err := p.parseSelectOrValues()
	if err != nil {
		return nil, err
	}

	var returningExpressions []nodes.ProjectionExpression
	if p.advanceIf(isType(TokenTypeReturning)) {
		returningExpressions, err = p.parseSelectExpressions()
		if err != nil {
			return nil, err
		}

		if returningExpressions == nil {
			returningExpressions = []nodes.ProjectionExpression{
				nodes.NewTableWildcardProjectionExpression(nameToken.Text),
			}
		}
	}

	return nodes.NewInsert(node, table, nameToken.Text, alias, columnNames, returningExpressions)
}

// selectOrValues := `SELECT` select
//                 | values
func (p *parser) parseSelectOrValues() (nodes.Node, error) {
	token := p.current()
	if p.advanceIf(isType(TokenTypeSelect)) {
		return p.parseSelect(token)
	}

	return p.parseValues()
}

// values := `VALUES` valuesList
func (p *parser) parseValues() (nodes.Node, error) {
	if _, err := p.mustAdvance(isType(TokenTypeValues)); err != nil {
		return nil, err
	}

	return p.parseValuesList()
}

// valuesList := ( `(` ( expression [, ...] ) `)` [, ...] )
func (p *parser) parseValuesList() (nodes.Node, error) {
	var allRows [][]interface{}
	for {
		if _, err := p.mustAdvance(isType(TokenTypeLeftParen)); err != nil {
			return nil, err
		}

		var values []interface{}
		for {
			expression, err := p.parseExpression(0)
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
		fields = append(fields, shared.NewField("", fmt.Sprintf("column%d", i+1), shared.TypeKindAny, false))
	}

	rows, err := shared.NewRows(fields)
	if err != nil {
		return nil, err
	}

	for _, values := range allRows {
		var err error
		rows, err = rows.AddValues(values)
		if err != nil {
			return nil, err
		}
	}

	table, err := nodes.NewTable("", rows)
	if err != nil {
		return nil, err
	}

	return nodes.NewData(table), nil
}

//
// Update statements

// update := ident [[`AS`] ident] `SET` ( ident `=` expression [, ...] ) where [`RETURNING` selectExpressions]
func (p *parser) parseUpdate(token Token) (nodes.Node, error) {
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

	if _, err := p.mustAdvance(isType(TokenTypeSet)); err != nil {
		return nil, err
	}

	var setExpressions []nodes.SetExpression
	for {
		var name string
		if p.advanceIf(isType(TokenTypeLeftParen)) {
			// TODO - implement
			panic("Multi-column sets unimplemented")
		} else {
			nameToken, err := p.mustAdvance(isType(TokenTypeIdent))
			if err != nil {
				return nil, err
			}

			name = nameToken.Text
		}

		if _, err := p.mustAdvance(isType(TokenTypeEquals)); err != nil {
			return nil, err
		}

		// TODO - or subselect
		expr, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		setExpressions = append(setExpressions, nodes.SetExpression{
			Name:       name,
			Expression: expr,
		})

		if !p.advanceIf(isType(TokenTypeComma)) {
			break
		}
	}

	// TODO - parse/support FROM

	whereExpression, hasWhere, err := p.parseWhereClause()
	if err != nil {
		return nil, err
	}

	var returningExpressions []nodes.ProjectionExpression
	if p.advanceIf(isType(TokenTypeReturning)) {
		returningExpressions, err = p.parseSelectExpressions()
		if err != nil {
			return nil, err
		}

		if returningExpressions == nil {
			returningExpressions = []nodes.ProjectionExpression{
				nodes.NewTableWildcardProjectionExpression(nameToken.Text),
			}
		}
	}

	node := nodes.NewData(table)
	if alias != "" {
		node = nodes.NewAlias(node, alias)
	}
	if hasWhere {
		node = nodes.NewFilter(node, whereExpression)
	}
	return nodes.NewUpdate(node, table, setExpressions, alias, returningExpressions)
}

//
// Delete statements

// delete := `FROM` ident [[`AS`] ident] where [`RETURNING` selectExpressions]
func (p *parser) parseDelete(token Token) (nodes.Node, error) {
	if _, err := p.mustAdvance(isType(TokenTypeFrom)); err != nil {
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

	// TODO - parse/support FROM

	whereExpression, hasWhere, err := p.parseWhereClause()
	if err != nil {
		return nil, err
	}

	var returningExpressions []nodes.ProjectionExpression
	if p.advanceIf(isType(TokenTypeReturning)) {
		returningExpressions, err = p.parseSelectExpressions()
		if err != nil {
			return nil, err
		}

		if returningExpressions == nil {
			returningExpressions = []nodes.ProjectionExpression{
				nodes.NewTableWildcardProjectionExpression(nameToken.Text),
			}
		}
	}

	node := nodes.NewData(table)
	if alias != "" {
		node = nodes.NewAlias(node, alias)
	}
	if hasWhere {
		node = nodes.NewFilter(node, whereExpression)
	}
	return nodes.NewDelete(node, table, alias, returningExpressions)
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
func (p *parser) parseNamedExpression(token Token) (expressions.Expression, error) {
	if !p.advanceIf(isType(TokenTypeDot)) {
		return expressions.NewNamed(shared.NewField("", token.Text, shared.TypeKindAny, false)), nil
	}

	qualifiedNameToken, err := p.mustAdvance(isType(TokenTypeIdent))
	if err != nil {
		return nil, err
	}

	return expressions.NewNamed(shared.NewField(token.Text, qualifiedNameToken.Text, shared.TypeKindAny, false)), nil
}

// numericLiteralExpression := number
func (p *parser) parseNumericLiteralExpression(token Token) (expressions.Expression, error) {
	value, err := strconv.Atoi(token.Text)
	if err != nil {
		return nil, err
	}

	return expressions.NewConstant(value), nil
}

// numericLiteralExpression := string
func (p *parser) parseStringLiteralExpression(token Token) (expressions.Expression, error) {
	return expressions.NewConstant(token.Text), nil
}

// numericLiteralExpression := true | false
func (p *parser) parseBooleanLiteralExpression(token Token) (expressions.Expression, error) {
	return expressions.NewConstant(token.Type == TokenTypeTrue), nil
}

// numericLiteralExpression := null
func (p *parser) parseNullLiteralExpression(token Token) (expressions.Expression, error) {
	return expressions.NewConstant(nil), nil
}

// parenthesizedExpression := `(` expression `)`
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
