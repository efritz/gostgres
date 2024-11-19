package projection

import (
	"fmt"
	"slices"
	"strings"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
)

type Projection struct {
	aliases         []ProjectedExpression
	projectedFields []fields.Field
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
	var projectedFields []fields.Field
	for _, field := range projectedExpressions {
		relationName := targetRelationName
		if relationName == "" {
			if named, ok := field.Expression.(expressions.NamedExpression); ok {
				relationName = named.Field().RelationName()
			}
		}

		projectedFields = append(projectedFields, fields.NewField(relationName, field.Alias, field.Expression.Type(), fields.NonInternalField))
	}

	return &Projection{
		aliases:         projectedExpressions,
		projectedFields: projectedFields,
	}, nil
}

func (p *Projection) String() string {
	fields := make([]string, 0, len(p.aliases))

	relationNames := map[string]struct{}{}
	for _, expression := range p.aliases {
		if named, ok := expression.Expression.(expressions.NamedExpression); ok {
			relationNames[named.Field().RelationName()] = struct{}{}
		}
	}

	for _, expression := range p.aliases {
		// TODO - simplify named expressions below top-level?
		if named, ok := expression.Expression.(expressions.NamedExpression); ok {
			name := named.Field().String()
			if len(relationNames) == 1 {
				name = named.Field().Name()
			}

			if named.Field().Name() == expression.Alias {
				fields = append(fields, name)
			} else {
				fields = append(fields, fmt.Sprintf("%s as %s", name, expression.Alias))
			}

			continue
		}

		fields = append(fields, expression.String())
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
