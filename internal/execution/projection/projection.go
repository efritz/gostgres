package projection

import (
	"slices"
	"strings"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/types"
)

type Projection struct {
	aliases         []ProjectedExpression
	projectedFields []fields.Field
}

func NewProjection(
	targetRelationName string,
	relationFields []fields.Field,
	projectionExpressions []ProjectionExpression,
	aliasedTables ...AliasedTable,
) (*Projection, error) {
	projectedExpressions, err := ExpandProjection(relationFields, projectionExpressions, aliasedTables...)
	if err != nil {
		return nil, err
	}

	var projectedFields []fields.Field
	for _, field := range projectedExpressions {
		relationName := targetRelationName
		if relationName == "" {
			if named, ok := field.Expression.(expressions.NamedExpression); ok {
				relationName = named.Field().RelationName()
			}
		}

		projectedFields = append(projectedFields, fields.NewField(relationName, field.Alias, types.TypeAny, fields.NonInternalField))
	}

	return &Projection{
		aliases:         projectedExpressions,
		projectedFields: projectedFields,
	}, nil
}

func (p *Projection) String() string {
	fields := make([]string, 0, len(p.aliases))
	for _, expression := range p.aliases {
		if named, ok := expression.Expression.(expressions.NamedExpression); ok && named.Field().Name() == expression.Alias {
			fields = append(fields, expression.Alias)
		} else {
			fields = append(fields, expression.String())
		}
	}

	return strings.Join(fields, ", ")
}

func (p *Projection) Aliases() []ProjectedExpression {
	return slices.Clone(p.aliases)
}

func (p *Projection) Fields() []fields.Field {
	return slices.Clone(p.projectedFields)
}

func (p *Projection) Optimize() {
	for i := range p.aliases {
		p.aliases[i].Expression = p.aliases[i].Expression.Fold()
	}
}
