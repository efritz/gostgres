package parsing

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/types"
	"github.com/efritz/gostgres/internal/syntax/ast"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// table := ident alias
func (p *parser) parseTable() (ast.TableDescription, error) {
	name, err := p.parseIdent()
	if err != nil {
		return ast.TableDescription{}, err
	}

	alias, _, err := p.parseAlias()
	if err != nil {
		return ast.TableDescription{}, err
	}

	return ast.TableDescription{
		Name:      name,
		AliasName: alias,
	}, nil
}

// from := `FROM` tableExpressions
func (p *parser) parseFrom() (node ast.TableExpressionDescription, _ error) {
	if _, err := p.mustAdvance(isType(tokens.TokenTypeFrom)); err != nil {
		return ast.TableExpressionDescription{}, err
	}

	tableExpressions, err := p.parseTableExpressions()
	if err != nil {
		return ast.TableExpressionDescription{}, err
	}

	return joinNodes(tableExpressions), nil
}

// tableExpressions := tableExpression [, ...]
func (p *parser) parseTableExpressions() ([]ast.TableExpressionDescription, error) {
	return parseCommaSeparatedList(p, p.parseTableExpression)
}

// tableExpression := aliasedBaseTableExpression [ `JOIN` joinTail [...] ]
func (p *parser) parseTableExpression() (ast.TableExpressionDescription, error) {
	node, err := p.parseAliasedBaseTableExpression()
	if err != nil {
		return ast.TableExpressionDescription{}, err
	}

	var joins []ast.Join
	for p.advanceIf(isType(tokens.TokenTypeJoin)) {
		join, err := p.parseJoin()
		if err != nil {
			return ast.TableExpressionDescription{}, err
		}

		joins = append(joins, join)
	}

	return ast.TableExpressionDescription{
		Base:  node,
		Joins: joins,
	}, nil
}

// aliasedBaseTableExpression := baseTableExpression [ tableAlias ]
func (p *parser) parseAliasedBaseTableExpression() (ast.AliasedBaseTableExpressionDescription, error) {
	baseTableExpression, err := p.parseBaseTableExpression()
	if err != nil {
		return ast.AliasedBaseTableExpressionDescription{}, err
	}

	alias, err := p.parseTableAlias()
	if err != nil {
		return ast.AliasedBaseTableExpressionDescription{}, err
	}

	return ast.AliasedBaseTableExpressionDescription{
		BaseTableExpression: baseTableExpression,
		Alias:               alias,
	}, nil
}

// baseTableExpression := ( `(` selectOrValues `)` ) | ( `(` tableExpression `)` ) | tableReference
func (p *parser) parseBaseTableExpression() (ast.BaseTableExpressionDescription, error) {
	if p.advanceIf(isType(tokens.TokenTypeLeftParen)) {
		if p.current().Type == tokens.TokenTypeSelect || p.current().Type == tokens.TokenTypeValues {
			baseTableExpression, err := p.parseSelectOrValues()
			if err != nil {
				return nil, err
			}

			if _, err := p.mustAdvance(isType(tokens.TokenTypeRightParen)); err != nil {
				return nil, err
			}

			return baseTableExpression, nil
		}

		baseTableExpression, err := p.parseTableExpression()
		if err != nil {
			return nil, err
		}

		if _, err := p.mustAdvance(isType(tokens.TokenTypeRightParen)); err != nil {
			return nil, err
		}

		return baseTableExpression, nil
	}

	return p.parseTableReference()
}

// tableReference := ident
func (p *parser) parseTableReference() (ast.BaseTableExpressionDescription, error) {
	nameToken, err := p.parseIdent()
	if err != nil {
		return ast.TableReference{}, err
	}

	return ast.TableReference{
		Name: nameToken,
	}, nil
}

// selectOrValues := ( `SELECT` selectTail ) | values
func (p *parser) parseSelectOrValues() (ast.BaseTableExpressionDescription, error) {
	if p.advanceIf(isType(tokens.TokenTypeSelect)) {
		return p.parseSelect()
	}

	return p.parseValues()
}

// values := `VALUES` ( `(` ( expression [, ... ] ) `)` [, ...] )
func (p *parser) parseValues() (*ast.ValuesBuilder, error) {
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
	builder := &ast.ValuesBuilder{
		RowFields:         rowFields,
		AllRowExpressions: allRowExpressions,
	}

	return builder, nil
}

// tableAlias := alias [ `(` ident [, ...] `)` ]
func (p *parser) parseTableAlias() (*ast.TableAlias, error) {
	alias, hasAlias, err := p.parseAlias()
	if err != nil {
		return nil, err
	}
	if !hasAlias {
		return nil, nil
	}

	columnNames, err := parseParenthesizedCommaSeparatedList(p, true, false, p.parseIdent)
	if err != nil {
		return nil, nil
	}

	return &ast.TableAlias{
		TableAlias:    alias,
		ColumnAliases: columnNames,
	}, nil
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
func (p *parser) parseJoin() (ast.Join, error) {
	table, err := p.parseTableExpression()
	if err != nil {
		return ast.Join{}, err
	}

	var condition impls.Expression
	if p.advanceIf(isType(tokens.TokenTypeOn)) {
		rawCondition, err := p.parseRootExpression()
		if err != nil {
			return ast.Join{}, err
		}

		condition = rawCondition
	}

	// TODO - support USING
	// TODO - support multiple join types: (natural, left, outer, full, cross, and combinations)
	return ast.Join{
		Table:     table,
		Condition: condition,
	}, nil
}

func joinNodes(expressions []ast.TableExpressionDescription) ast.TableExpressionDescription {
	if len(expressions) == 0 {
		return ast.TableExpressionDescription{}
	}

	base := ast.AliasedBaseTableExpressionDescription{
		BaseTableExpression: expressions[0],
		Alias:               nil,
	}

	var joins []ast.Join
	for _, right := range expressions[1:] {
		joins = append(joins, ast.Join{
			Table:     right,
			Condition: nil,
		})
	}

	return ast.TableExpressionDescription{
		Base:  base,
		Joins: joins,
	}
}
