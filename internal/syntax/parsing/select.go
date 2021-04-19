package parsing

import (
	"fmt"
	"strconv"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/nodes"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// select := selectExpressions from where order limit offset
func (p *parser) parseSelect(token tokens.Token) (nodes.Node, error) {
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
	if p.advanceIf(isType(tokens.TokenTypeAsterisk)) {
		return []nodes.ProjectionExpression{nodes.NewWildcardProjectionExpression()}, nil
	}

	for {
		aliasedExpression, err := p.parseSelectExpression()
		if err != nil {
			return nil, err
		}

		aliasedExpressions = append(aliasedExpressions, aliasedExpression)

		if !p.advanceIf(isType(tokens.TokenTypeComma)) {
			break
		}
	}

	return aliasedExpressions, nil
}

// selectExpression := ident `.` `*`
//                   | expression columnAlias
func (p *parser) parseSelectExpression() (nodes.ProjectionExpression, error) {
	nameToken := p.current()
	if p.advanceIf(isType(tokens.TokenTypeIdent), isType(tokens.TokenTypeDot), isType(tokens.TokenTypeAsterisk)) {
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

// columnAlias := alias
func (p *parser) parseColumnAlias(expression expressions.Expression) (string, error) {
	alias, ok, err := p.parseAlias()
	if err != nil {
		return "", err
	}
	if ok {
		return alias, nil
	}

	type named interface {
		Name() string
	}
	if named, ok := expression.(named); ok {
		return named.Name(), nil
	}

	return "?column?", nil
}

// alias := [[`AS`] ident]
func (p *parser) parseAlias() (string, bool, error) {
	if p.advanceIf(isType(tokens.TokenTypeAs)) {
		aliasToken, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
		if err != nil {
			return "", false, err
		}

		return aliasToken.Text, true, nil
	}

	nameToken := p.current()
	if p.advanceIf(isType(tokens.TokenTypeIdent)) {
		return nameToken.Text, true, nil
	}

	return "", false, nil
}

// table := ident alias
func (p *parser) parseTable() (*nodes.Table, string, string, error) {
	nameToken, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
	if err != nil {
		return nil, "", "", err
	}

	table, ok := p.tables[nameToken.Text]
	if !ok {
		return nil, "", "", fmt.Errorf("unknown table %s", nameToken.Text)
	}

	alias, _, err := p.parseAlias()
	if err != nil {
		return nil, "", "", err
	}

	return table, nameToken.Text, alias, nil
}

// returning := [`RETURNING` selectExpressions]
func (p *parser) parseReturning(name string) (returningExpressions []nodes.ProjectionExpression, err error) {
	if !p.advanceIf(isType(tokens.TokenTypeReturning)) {
		return nil, nil
	}

	returningExpressions, err = p.parseSelectExpressions()
	if err != nil {
		return nil, err
	}
	if returningExpressions != nil {
		return returningExpressions, nil
	}

	return []nodes.ProjectionExpression{nodes.NewTableWildcardProjectionExpression(name)}, nil
}

// from := `FROM` tableExpressions
func (p *parser) parseFrom() (node nodes.Node, _ error) {
	if _, err := p.mustAdvance(isType(tokens.TokenTypeFrom)); err != nil {
		return nil, err
	}

	tableExpressions, err := p.parseTableExpressions()
	if err != nil {
		return nil, err
	}

	return joinNodes(tableExpressions), nil
}

// tableExpressions := tableExpression [, ...]
func (p *parser) parseTableExpressions() ([]nodes.Node, error) {
	var tableExpressions []nodes.Node
	for {
		tableExpression, err := p.parseTableExpression()
		if err != nil {
			return nil, err
		}

		tableExpressions = append(tableExpressions, tableExpression)

		if !p.advanceIf(isType(tokens.TokenTypeComma)) {
			break
		}
	}

	return tableExpressions, nil
}

// tableExpression := baseTableExpression [( `JOIN` join [...] )]
func (p *parser) parseTableExpression() (nodes.Node, error) {
	node, err := p.parseBaseTableExpression()
	if err != nil {
		return nil, err
	}

	for p.advanceIf(isType(tokens.TokenTypeJoin)) {
		node, err = p.parseJoin(node)
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

// join := tableExpression [`ON` expression]
func (p *parser) parseJoin(node nodes.Node) (nodes.Node, error) {
	right, err := p.parseTableExpression()
	if err != nil {
		return nil, err
	}

	var condition expressions.Expression
	if p.advanceIf(isType(tokens.TokenTypeOn)) {
		rawCondition, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		condition = rawCondition
	}

	return nodes.NewJoin(node, right, condition), nil
}

// baseTableExpression := tableReference alias
//                      | `(` ( selectOrValues | tableExpression ) `)` alias
func (p *parser) parseBaseTableExpression() (nodes.Node, error) {
	expectParen := false
	requireAlias := false
	parseFunc := p.parseTableReference

	if p.advanceIf(isType(tokens.TokenTypeLeftParen)) {
		if p.current().Type == tokens.TokenTypeSelect || p.current().Type == tokens.TokenTypeValues {
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
		if _, err := p.mustAdvance(isType(tokens.TokenTypeRightParen)); err != nil {
			return nil, err
		}
	}

	alias, hasAlias, err := p.parseAlias()
	if err != nil {
		return nil, err
	}
	if hasAlias {
		node = nodes.NewAlias(node, alias)
	} else if requireAlias {
		return nil, fmt.Errorf("expected subselect alias (near %s)", p.current().Text)
	}

	return node, nil
}

// tableReference := ident
func (p *parser) parseTableReference() (nodes.Node, error) {
	nameToken, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
	if err != nil {
		return nil, err
	}

	table, ok := p.tables[nameToken.Text]
	if !ok {
		return nil, fmt.Errorf("unknown table %s", nameToken.Text)
	}

	return nodes.NewData(table), nil
}

// where := [`WHERE` expression]
func (p *parser) parseWhereClause() (expressions.Expression, bool, error) {
	if !p.advanceIf(isType(tokens.TokenTypeWhere)) {
		return nil, false, nil
	}

	whereExpression, err := p.parseExpression(0)
	if err != nil {
		return nil, false, err
	}

	return whereExpression, true, nil
}

// where := [`ORDER` `BY` ( expression [( `ASC` | `DESC` )] [, ...] )]
func (p *parser) parseOrderByClause() (nodes.OrderExpression, bool, error) {
	if !p.advanceIf(isType(tokens.TokenTypeOrder)) {
		return nil, false, nil
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeBy)); err != nil {
		return nil, false, err
	}

	var expressions []nodes.FieldExpression
	for {
		orderExpression, err := p.parseExpression(0)
		if err != nil {
			return nil, false, err
		}

		reverse := false
		if !p.advanceIf(isType(tokens.TokenTypeAscending)) {
			if p.advanceIf(isType(tokens.TokenTypeDescending)) {
				reverse = true
			}
		}

		expressions = append(expressions, nodes.FieldExpression{
			Expression: orderExpression,
			Reverse:    reverse,
		})

		if !p.advanceIf(isType(tokens.TokenTypeComma)) {
			break
		}
	}

	return nodes.NewOrderExpression(expressions), true, nil
}

// limit := [`LIMIT` expression]
func (p *parser) parseLimitClause() (int, bool, error) {
	if !p.advanceIf(isType(tokens.TokenTypeLimit)) {
		return 0, false, nil
	}

	limitToken, err := p.mustAdvance(isType(tokens.TokenTypeNumber))
	if err != nil {
		return 0, false, err
	}

	limitValue, err := strconv.Atoi(limitToken.Text)
	return limitValue, true, err
}

// offset := [`LIMIT` expression]
func (p *parser) parseOffsetClause() (int, bool, error) {
	if !p.advanceIf(isType(tokens.TokenTypeOffset)) {
		return 0, false, nil
	}

	limitToken, err := p.mustAdvance(isType(tokens.TokenTypeNumber))
	if err != nil {
		return 0, false, err
	}

	limitValue, err := strconv.Atoi(limitToken.Text)
	return limitValue, true, err
}

// selectOrValues := `SELECT` select
//                 | values
func (p *parser) parseSelectOrValues() (nodes.Node, error) {
	token := p.current()
	if p.advanceIf(isType(tokens.TokenTypeSelect)) {
		return p.parseSelect(token)
	}

	return p.parseValues()
}

// values := `VALUES` valuesList
func (p *parser) parseValues() (nodes.Node, error) {
	if _, err := p.mustAdvance(isType(tokens.TokenTypeValues)); err != nil {
		return nil, err
	}

	return p.parseValuesList()
}

// valuesList := ( parenthesizedExpressionList [, ...] )
func (p *parser) parseValuesList() (nodes.Node, error) {
	var allRows [][]interface{}
	for {
		expressions, err := p.parseParenthesizedExpressionList()
		if err != nil {
			return nil, err
		}

		values := make([]interface{}, 0, len(expressions))
		for _, expression := range expressions {
			value, err := expression.ValueFrom(shared.Row{})
			if err != nil {
				return nil, err
			}

			values = append(values, value)
		}

		allRows = append(allRows, values)

		if !p.advanceIf(isType(tokens.TokenTypeComma)) {
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

// parenthesizedExpressionList := `(` ( expression [, ... ] ) `)`
func (p *parser) parseParenthesizedExpressionList() ([]expressions.Expression, error) {
	if _, err := p.mustAdvance(isType(tokens.TokenTypeLeftParen)); err != nil {
		return nil, err
	}

	var expressions []expressions.Expression
	for {
		expression, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, expression)

		if !p.advanceIf(isType(tokens.TokenTypeComma)) {
			break
		}
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeRightParen)); err != nil {
		return nil, err
	}

	return expressions, nil
}
