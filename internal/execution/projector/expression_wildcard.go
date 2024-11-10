package projector

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
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

func (p wildcardProjectionExpression) Expand(fields []fields.Field) (projections []AliasProjectionExpression, _ error) {
	for _, field := range fields {
		if field.Internal() {
			continue
		}

		projections = append(projections, AliasProjectionExpression{
			Alias:      field.Name(),
			Expression: expressions.NewNamed(field),
		})
	}

	return projections, nil
}

func (p wildcardProjectionExpression) Map(f func(impls.Expression) (impls.Expression, error)) (ProjectionExpression, error) {
	return p, nil
}
