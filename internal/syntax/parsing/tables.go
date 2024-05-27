package parsing

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/access"
	"github.com/efritz/gostgres/internal/execution/queries/alias"
	"github.com/efritz/gostgres/internal/execution/queries/joins"
	"github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/types"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// table := ident alias
func (p *parser) parseTable() (impls.Table, string, string, error) {
	name, err := p.parseIdent()
	if err != nil {
		return nil, "", "", err
	}

	table, ok := p.tables.Get(name)
	if !ok {
		return nil, "", "", fmt.Errorf("unknown table %s", name)
	}

	alias, _, err := p.parseAlias()
	if err != nil {
		return nil, "", "", err
	}

	return table, name, alias, nil
}

// from := `FROM` tableExpressions
func (p *parser) parseFrom() (node queries.Node, _ error) {
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
func (p *parser) parseTableExpressions() ([]queries.Node, error) {
	return parseCommaSeparatedList(p, p.parseTableExpression)
}

// tableExpression := baseTableExpression [ `JOIN` joinTail [...] ]
func (p *parser) parseTableExpression() (queries.Node, error) {
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

// baseTableExpression := ( tableReference [ tableAlias ] ) | ( `(` selectOrValues `)` tableAlias ) | ( `(` tableExpression `)` [ tableAlias ] )
func (p *parser) parseBaseTableExpression() (queries.Node, error) {
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
			var fields []fields.Field
			for _, f := range node.Fields() {
				if !f.Internal() {
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
func (p *parser) parseTableReference() (queries.Node, error) {
	nameToken, err := p.parseIdent()
	if err != nil {
		return nil, err
	}

	table, ok := p.tables.Get(nameToken)
	if !ok {
		return nil, fmt.Errorf("unknown table %s", nameToken)
	}

	return access.NewAccess(table), nil
}

// selectOrValues := ( `SELECT` selectTail ) | values
func (p *parser) parseSelectOrValues() (queries.Node, error) {
	token := p.current()
	if p.advanceIf(isType(tokens.TokenTypeSelect)) {
		return p.parseSelect(token)
	}

	return p.parseValues()
}

func (p *parser) parseSelectOrValuesBuilder() (Builder, error) {
	if p.advanceIf(isType(tokens.TokenTypeSelect)) {
		return p.parseSelectBuilder()
	}

	return p.parseValuesBuilder()
}

// values := `VALUES` ( `(` ( expression [, ... ] ) `)` [, ...] )
func (p *parser) parseValues() (queries.Node, error) {
	builder, err := p.parseValuesBuilder()
	if err != nil {
		return nil, err
	}

	return builder.Build()
}

func (p *parser) parseValuesBuilder() (Builder, error) {
	if _, err := p.mustAdvance(isType(tokens.TokenTypeValues)); err != nil {
		return nil, err
	}

	allRowExpressions, err := parseCommaSeparatedList(p, func() ([]impls.Expression, error) {
		return parseParenthesizedCommaSeparatedList(p, false, false, p.parseRootExpression)
	})
	if err != nil {
		return nil, err
	}

	rowFields := make([]fields.Field, 0, len(allRowExpressions[0]))
	for i := range allRowExpressions[0] {
		rowFields = append(rowFields, fields.NewField("", fmt.Sprintf("column%d", i+1), types.TypeAny))
	}

	// TODO - support `DEFAULT` expressions
	builder := &ValuesBuilder{
		rowFields:         rowFields,
		allRowExpressions: allRowExpressions,
	}

	return builder, nil
}

// tableAlias := alias [ `(` ident [, ...] `)` ]
func (p *parser) parseTableAlias() (string, []string, bool, error) {
	alias, hasAlias, err := p.parseAlias()
	if err != nil {
		return "", nil, false, err
	}
	if !hasAlias {
		return "", nil, false, nil
	}

	columnNames, err := parseParenthesizedCommaSeparatedList(p, true, false, p.parseIdent)
	if err != nil {
		return "", nil, false, nil
	}

	return alias, columnNames, true, nil
}

// alias := [ [ `AS` ] ident ]
func (p *parser) parseAlias() (string, bool, error) {
	if p.advanceIf(isType(tokens.TokenTypeAs)) {
		alias, err := p.parseIdent()
		if err != nil {
			return "", false, err
		}

		return alias, true, nil
	}

	nameToken := p.current()
	if p.advanceIf(isType(tokens.TokenTypeIdent)) {
		return nameToken.Text, true, nil
	}

	return "", false, nil
}

// joinTail := tableExpression [`ON` expression]
func (p *parser) parseJoin(node queries.Node) (queries.Node, error) {
	right, err := p.parseTableExpression()
	if err != nil {
		return nil, err
	}

	var condition impls.Expression
	if p.advanceIf(isType(tokens.TokenTypeOn)) {
		rawCondition, err := p.parseRootExpression()
		if err != nil {
			return nil, err
		}

		condition = rawCondition
	}

	// TODO - support USING
	// TODO - support multiple join types: (natural, left, outer, full, cross, and combinations)
	return joins.NewJoin(node, right, condition), nil
}

func joinNodes(expressions []queries.Node) queries.Node {
	if len(expressions) == 0 {
		return nil
	}

	left := expressions[0]
	for _, right := range expressions[1:] {
		left = joins.NewJoin(left, right, nil)
	}

	return left
}
