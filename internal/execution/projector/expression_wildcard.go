package projector

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
)

type wildcardProjectionExpression struct{}

func NewWildcardProjectionExpression() ProjectionExpression {
	return wildcardProjectionExpression{}
}

func (p wildcardProjectionExpression) String() string {
	return "*"
}

func (p wildcardProjectionExpression) Dealias(name string, fields []fields.Field, alias string) ProjectionExpression {
	return p
}

func (p wildcardProjectionExpression) Expand(fields []fields.Field) (projections []aliasProjectionExpression, _ error) {
	for _, field := range fields {
		if field.Internal() {
			continue
		}

		projections = append(projections, aliasProjectionExpression{
			alias:      field.Name(),
			expression: expressions.NewNamed(field),
		})
	}

	return projections, nil
}
