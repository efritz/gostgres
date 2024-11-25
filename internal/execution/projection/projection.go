package projection

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/types"
)

type Projection struct {
	targetRelationName string
	aliases            []ProjectedExpression
	projectedFields    []fields.Field
}

func NewProjectionFromProjectionExpressions(
	targetRelationName string,
	relationFields []fields.Field,
	projectionExpressions []ProjectionExpression,
	aliasedTables ...AliasedTable,
) (*Projection, error) {
	projectedExpressions, err := ExpandProjection(relationFields, projectionExpressions, aliasedTables...)
	if err != nil {
		return nil, err
	}

	return NewProjectionFromProjectedExpressions(targetRelationName, projectedExpressions)
}

func NewProjectionFromProjectedExpressions(
	targetRelationName string,
	projectedExpressions []ProjectedExpression,
) (*Projection, error) {
	return &Projection{
		targetRelationName: targetRelationName,
		aliases:            projectedExpressions,
		projectedFields:    fieldsFromProjectedExpressions(targetRelationName, projectedExpressions),
	}, nil
}

func (p *Projection) String() string {
	suffix := ""
	if p.targetRelationName != "" {
		suffix = fmt.Sprintf(" into %s.*", p.targetRelationName)
	}

	return fmt.Sprintf("{%s}%s", serializeProjectedExpressions(p.aliases), suffix)
}

func (p *Projection) Aliases() []ProjectedExpression {
	return slices.Clone(p.aliases)
}

func (p *Projection) Fields() []fields.Field {
	return slices.Clone(p.projectedFields)
}

func (p *Projection) Optimize(ctx impls.OptimizationContext) {
	for i := range p.aliases {
		p.aliases[i].Expression = p.aliases[i].Expression.Fold()
	}
}

func (p *Projection) ProjectExpression(expression impls.Expression) impls.Expression {
	for i, alias := range p.aliases {
		if named, ok := alias.Expression.(expressions.NamedExpression); ok {
			expression = Alias(expression, named.Field(), expressions.NewNamed(p.projectedFields[i]))
		}
	}

	return expression
}

func (p *Projection) DeprojectExpression(expression impls.Expression) impls.Expression {
	for i, alias := range p.aliases {
		field := fields.NewField(p.projectedFields[i].RelationName(), alias.Alias, types.TypeAny, fields.NonInternalField)
		expression = Alias(expression, field, alias.Expression)
	}

	return expression
}
