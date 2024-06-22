package projector

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
)

type tableWildcardProjectionExpression struct {
	relationName string
}

func NewTableWildcardProjectionExpression(relationName string) ProjectionExpression {
	return tableWildcardProjectionExpression{
		relationName: relationName,
	}
}

func (p tableWildcardProjectionExpression) String() string {
	return fmt.Sprintf("%s.*", p.relationName)
}

func (p tableWildcardProjectionExpression) Dealias(name string, fields []fields.Field, alias string) ProjectionExpression {
	if p.relationName == alias {
		return tableWildcardProjectionExpression{
			relationName: name,
		}
	}

	return p
}

func (p tableWildcardProjectionExpression) Expand(fields []fields.Field) (projections []aliasProjectionExpression, _ error) {
	matched := false
	for _, field := range fields {
		if field.Internal() || field.RelationName() != p.relationName {
			continue
		}

		matched = true
		projections = append(projections, aliasProjectionExpression{
			expression:   expressions.NewNamed(field),
			relationName: p.relationName,
			aliasName:    field.Name(),
		})
	}

	if !matched {
		return nil, fmt.Errorf("unknown relation %q", p.relationName)
	}

	return projections, nil
}
