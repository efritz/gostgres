package projector

import "github.com/efritz/gostgres/internal/shared/fields"

type ProjectionExpression interface {
	ProjectionExpression1
	ProjectionExpression2
}

type ProjectionExpression1 interface {
	Dealias(name string, fields []fields.Field, alias string) ProjectionExpression
}

type ProjectionExpression2 interface {
	Expand(fields []fields.Field) ([]ProjectedExpression, error)
}

func ExpandProjection(fields []fields.Field, expressions []ProjectionExpression) ([]ProjectedExpression, error) {
	aliases := make([]ProjectedExpression, 0, len(fields))
	for _, expression := range expressions {
		as, err := expression.Expand(fields)
		if err != nil {
			return nil, err
		}

		aliases = append(aliases, as...)
	}

	return aliases, nil
}
