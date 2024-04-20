package parsing

import (
	"fmt"
	"strconv"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/nodes"
	"github.com/efritz/gostgres/internal/queries/access"
	"github.com/efritz/gostgres/internal/queries/alias"
	"github.com/efritz/gostgres/internal/queries/combination"
	"github.com/efritz/gostgres/internal/queries/filter"
	"github.com/efritz/gostgres/internal/queries/joins"
	"github.com/efritz/gostgres/internal/queries/limit"
	"github.com/efritz/gostgres/internal/queries/order"
	"github.com/efritz/gostgres/internal/queries/projection"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/tokens"
	"github.com/efritz/gostgres/internal/table"
)

// select := simpleSelect orderby limit offset
func (p *parser) parseSelect(token tokens.Token) (nodes.Node, error) {
	selectNode, err := p.parseSimpleSelect(token)
	if err != nil {
		return nil, err
	}

	orderExpression, hasOrder, err := p.parseOrderBy()
	if err != nil {
		return nil, err
	}
	limitValue, hasLimit, err := p.parseLimit()
	if err != nil {
		return nil, err
	}
	offsetValue, hasOffset, err := p.parseOffset()
	if err != nil {
		return nil, err
	}

	// TODO - support [ FETCH { FIRST | NEXT } [ count ] { ROW | ROWS } ONLY ]
	// TODO - support [ FOR { UPDATE | SHARE } [ OF table_name [, ...] ] [ NOWAIT ] [...] ]

	node := selectNode.node
	if hasOrder {
		node = order.NewOrder(node, orderExpression)
	}
	if hasOffset {
		node = limit.NewOffset(node, offsetValue)
	}
	if hasLimit {
		node = limit.NewLimit(node, limitValue)
	}

	node, err = projection.NewProjection(node, selectNode.selectExpressions)
	if err != nil {
		return nil, err
	}

	return node, nil
}

type selectNode struct {
	node              nodes.Node
	selectExpressions []projection.ProjectionExpression
}

// simpleSelect := selectExpressions from where combinedQuery
func (p *parser) parseSimpleSelect(token tokens.Token) (selectNode, error) {
	// TODO - support [ `ALL` | `DISTINCT` [ `ON` ( expression [, ...] ) ] ]

	selectExpressions, err := p.parseSelectExpressions()
	if err != nil {
		return selectNode{}, err
	}

	// TODO - make from optional

	node, err := p.parseFrom()
	if err != nil {
		return selectNode{}, err
	}

	whereExpression, hasWhere, err := p.parseWhere()
	if err != nil {
		return selectNode{}, err
	}
	if hasWhere {
		node = filter.NewFilter(node, whereExpression)
	}

	// TODO - support [ `GROUP` `BY` expression [, ...] ]
	// TODO - support [ `HAVING` condition [, ...] ]
	// TODO - support [ `WINDOW` window_name `AS` ( window_definition ) [, ...] ]

	if p.current().Type != tokens.TokenTypeUnion && p.current().Type != tokens.TokenTypeIntersect && p.current().Type != tokens.TokenTypeExcept {
		return selectNode{
			node:              node,
			selectExpressions: selectExpressions,
		}, nil
	}

	node, err = projection.NewProjection(node, selectExpressions)
	if err != nil {
		return selectNode{}, err
	}
	node, err = p.parseCombinedQuery(node)
	if err != nil {
		return selectNode{}, err
	}

	return selectNode{
		node:              node,
		selectExpressions: []projection.ProjectionExpression{projection.NewWildcardProjectionExpression()},
	}, nil
}

// combinedQuery := [ ( `UNION` | `INTERSECT` | `EXCEPT` ) [( `ALL` | `DISTINCT` )] combinationTarget [, ...] ]
func (p *parser) parseCombinedQuery(node nodes.Node) (nodes.Node, error) {
	for {
		var factory func(left, right nodes.Node, distinct bool) (nodes.Node, error)
		if p.advanceIf(isType(tokens.TokenTypeUnion)) {
			factory = combination.NewUnion
		} else if p.advanceIf(isType(tokens.TokenTypeIntersect)) {
			factory = combination.NewIntersect
		} else if p.advanceIf(isType(tokens.TokenTypeExcept)) {
			factory = combination.NewExcept
		} else {
			break
		}

		distinct := true
		if p.advanceIf(isType(tokens.TokenTypeDistinct)) {
			// token is explicitly supplying the default
		} else if p.advanceIf(isType(tokens.TokenTypeAll)) {
			distinct = false
		}

		unionTarget, err := p.parseCombinationTarget()
		if err != nil {
			return nil, err
		}

		node, err = factory(node, unionTarget, distinct)
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

// combinationTarget := simpleSelect
//
//	| `(` selectOrValues `)`
func (p *parser) parseCombinationTarget() (nodes.Node, error) {
	expectParen := false
	var parseFunc func() (nodes.Node, error)

	if p.advanceIf(isType(tokens.TokenTypeLeftParen)) {
		expectParen = true
		parseFunc = p.parseSelectOrValues
	} else {
		parseFunc = func() (nodes.Node, error) {
			token, err := p.mustAdvance(isType(tokens.TokenTypeSelect))
			if err != nil {
				return nil, err
			}

			selectNode, err := p.parseSimpleSelect(token)
			if err != nil {
				return nil, err
			}

			return projection.NewProjection(selectNode.node, selectNode.selectExpressions)
		}
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

	return node, nil
}

// selectExpressions := `*`
//
//	| selectExpression [, ...]
func (p *parser) parseSelectExpressions() (aliasedExpressions []projection.ProjectionExpression, _ error) {
	if p.advanceIf(isType(tokens.TokenTypeAsterisk)) {
		return []projection.ProjectionExpression{projection.NewWildcardProjectionExpression()}, nil
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
//
//	| expression alias
func (p *parser) parseSelectExpression() (projection.ProjectionExpression, error) {
	nameToken := p.current()
	if p.advanceIf(isType(tokens.TokenTypeIdent), isType(tokens.TokenTypeDot), isType(tokens.TokenTypeAsterisk)) {
		return projection.NewTableWildcardProjectionExpression(nameToken.Text), nil
	}

	expression, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}

	type named interface {
		Name() string
	}
	var alias string
	if value, ok, err := p.parseAlias(); err != nil {
		return nil, err
	} else if ok {
		alias = value
	} else if named, ok := expression.(named); ok {
		alias = named.Name()
	} else {
		alias = "?column?"
	}

	return projection.NewAliasProjectionExpression(expression, alias), nil
}

// tableAlias := alias columnNames
func (p *parser) parseTableAlias() (string, []string, bool, error) {
	alias, hasAlias, err := p.parseAlias()
	if err != nil {
		return "", nil, false, err
	}
	if !hasAlias {
		return "", nil, false, nil
	}

	columnNames, err := p.parseColumnNames()
	if err != nil {
		return "", nil, false, nil
	}

	return alias, columnNames, true, nil
}

// columnNames := [ `(` ident [, ...] `)` ]
func (p *parser) parseColumnNames() ([]string, error) {
	if p.current().Type != tokens.TokenTypeLeftParen || p.peek(1).Type != tokens.TokenTypeIdent {
		return nil, nil
	}

	p.advance()

	var columnNames []string
	for {
		nameToken, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
		if err != nil {
			return nil, err
		}

		columnNames = append(columnNames, nameToken.Text)

		if !p.advanceIf(isType(tokens.TokenTypeComma)) {
			break
		}
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeRightParen)); err != nil {
		return nil, err
	}

	return columnNames, nil
}

// alias := [[`AS`] ident]
func (p *parser) parseAlias() (string, bool, error) {
	return p.parseAliasPrefix()
}

func (p *parser) parseAliasPrefix() (string, bool, error) {
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
func (p *parser) parseTable() (*table.Table, string, string, error) {
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
func (p *parser) parseReturning(name string) (returningExpressions []projection.ProjectionExpression, err error) {
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

	return []projection.ProjectionExpression{projection.NewTableWildcardProjectionExpression(name)}, nil
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

	// TODO - support multiple join types
	// [ NATURAL ] ( [ INNER ] | LEFT [ OUTER ] | RIGHT [ OUTER ] | FULL [ OUTER ] | CROSS ) JOIN

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

	// TODO - support [ `USING` ( ident [, ...] ) ]

	var condition expressions.Expression
	if p.advanceIf(isType(tokens.TokenTypeOn)) {
		rawCondition, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		condition = rawCondition
	}

	return joins.NewJoin(node, right, condition), nil
}

// baseTableExpression := tableReference tableAlias
//
//	| `(` ( selectOrValues | tableExpression ) `)` tableAlias
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

	aliasName, columnNames, hasAlias, err := p.parseTableAlias()
	if err != nil {
		return nil, err
	}
	if hasAlias {
		node = alias.NewAlias(node, aliasName)

		if len(columnNames) > 0 {
			var fields []shared.Field
			for _, f := range node.Fields() {
				if !f.Internal {
					fields = append(fields, f)
				}
			}

			if len(columnNames) != len(fields) {
				return nil, fmt.Errorf("wrong number of fields in alias")
			}

			projectionExpressions := make([]projection.ProjectionExpression, 0, len(fields))
			for i, field := range fields {
				projectionExpressions = append(projectionExpressions, projection.NewAliasProjectionExpression(expressions.NewNamed(field), columnNames[i]))
			}

			node, err = projection.NewProjection(node, projectionExpressions)
			if err != nil {
				return nil, err
			}
		}
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

	return access.NewData(table), nil
}

// where := [`WHERE` expression]
func (p *parser) parseWhere() (expressions.Expression, bool, error) {
	if !p.advanceIf(isType(tokens.TokenTypeWhere)) {
		return nil, false, nil
	}

	whereExpression, err := p.parseExpression(0)
	if err != nil {
		return nil, false, err
	}

	return whereExpression, true, nil
}

// orderby := [`ORDER` `BY` ( expression [( `ASC` | `DESC` )] [, ...] )]
func (p *parser) parseOrderBy() (expressions.OrderExpression, bool, error) {
	if !p.advanceIf(isType(tokens.TokenTypeOrder)) {
		return nil, false, nil
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeBy)); err != nil {
		return nil, false, err
	}

	// TODO - support `USING` operator
	// TODO - support [`NULLS` ( `FIRST` | `LAST` )]

	var orderExpressions []expressions.ExpressionWithDirection
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

		orderExpressions = append(orderExpressions, expressions.ExpressionWithDirection{
			Expression: orderExpression,
			Reverse:    reverse,
		})

		if !p.advanceIf(isType(tokens.TokenTypeComma)) {
			break
		}
	}

	return expressions.NewOrderExpression(orderExpressions), true, nil
}

// limit := [`LIMIT` expression]
func (p *parser) parseLimit() (int, bool, error) {
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
func (p *parser) parseOffset() (int, bool, error) {
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
//
//	| values
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

	// TODO - support `DEFAULT` expressions
	return p.parseValuesList()
}

// valuesList := ( parenthesizedExpressionList [, ...] )
func (p *parser) parseValuesList() (nodes.Node, error) {
	var allRows [][]any
	for {
		expressions, err := p.parseParenthesizedExpressionList()
		if err != nil {
			return nil, err
		}

		values := make([]any, 0, len(expressions))
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

	table := table.NewTable("", rows.Fields)

	for i := 0; i < rows.Size(); i++ {
		if _, err := table.Insert(rows.Row(i)); err != nil {
			return nil, err
		}
	}

	return access.NewData(table), nil
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
