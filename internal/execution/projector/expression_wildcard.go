package projector

import (
	"github.com/efritz/gostgres/internal/shared/fields"
)

type wildcardProjectionExpression struct{}

var _ ProjectionExpression1 = &wildcardProjectionExpression{}
var _ ProjectionExpression2 = &wildcardProjectionExpression{}

func NewWildcardProjectionExpression() ProjectionExpression {
	return wildcardProjectionExpression{}
}

func (p wildcardProjectionExpression) String() string {
	return "*"
}

func (p wildcardProjectionExpression) Dealias(name string, fields []fields.Field, alias string) ProjectionExpression {
	return p
}

func (p wildcardProjectionExpression) Expand(fields []fields.Field) (projections []ProjectedExpression, _ error) {
	for _, field := range fields {
		if field.Internal() {
			continue
		}

		projections = append(projections, NewProjectedExpressionFromField(field))
	}

	return projections, nil
}
