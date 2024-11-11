package projector

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type ProjectionExpression interface {
	Expand(fields []fields.Field, aliasedTables ...AliasedTable) ([]ProjectedExpression, error)
}

type AliasedTable struct {
	TableName string
	Alias     string
}

func ExpandProjection(fields []fields.Field, expressions []ProjectionExpression, aliasedTables ...AliasedTable) ([]ProjectedExpression, error) {
	aliases := make([]ProjectedExpression, 0, len(fields))
	for _, expression := range expressions {
		as, err := expression.Expand(fields, aliasedTables...)
		if err != nil {
			return nil, err
		}

		aliases = append(aliases, as...)
	}

	return aliases, nil
}

//
//

type wildcardProjectionExpression struct{}

func NewWildcardProjectionExpression() ProjectionExpression {
	return wildcardProjectionExpression{}
}

func (p wildcardProjectionExpression) Expand(fields []fields.Field, aliasedTables ...AliasedTable) (projections []ProjectedExpression, _ error) {
	for _, field := range fields {
		if field.Internal() {
			continue
		}

		projections = append(projections, NewProjectedExpressionFromField(field))
	}

	return projections, nil
}

//
//

type tableWildcardProjectionExpression struct {
	relationName string
}

func NewTableWildcardProjectionExpression(relationName string) ProjectionExpression {
	return tableWildcardProjectionExpression{
		relationName: relationName,
	}
}

func (p tableWildcardProjectionExpression) Expand(fields []fields.Field, aliasedTables ...AliasedTable) (projections []ProjectedExpression, _ error) {
	name := p.relationName
	for _, aliasedTable := range aliasedTables {
		if p.relationName == aliasedTable.Alias {
			name = aliasedTable.TableName
		}
	}

	matched := false
	for _, field := range fields {
		if field.Internal() || field.RelationName() != name {
			continue
		}

		matched = true
		projections = append(projections, NewProjectedExpressionFromField(field))
	}

	if !matched {
		return nil, fmt.Errorf("unknown relation %q", name)
	}

	return projections, nil
}

//
//

type aliasedExpression struct {
	expression impls.Expression
	alias      string
}

func NewAliasedExpression(expression impls.Expression, alias string) ProjectionExpression {
	return aliasedExpression{
		expression: expression,
		alias:      alias,
	}
}

func NewAliasedExpressionFromField(field fields.Field) ProjectionExpression {
	return NewAliasedExpression(expressions.NewNamed(field), field.Name())
}

func (p aliasedExpression) Expand(fields []fields.Field, aliasedTables ...AliasedTable) ([]ProjectedExpression, error) {
	expression := p.expression
	for _, table := range aliasedTables {
		for _, field := range fields {
			expression = Alias(expression, field.WithRelationName(table.Alias), expressions.NewNamed(field))
		}
	}

	return []ProjectedExpression{NewProjectedExpression(expression, p.alias)}, nil
}
