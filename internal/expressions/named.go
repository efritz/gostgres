package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type namedExpression struct {
	relationName string
	name         string
}

var _ Expression = &namedExpression{}

func NewNamed(relationName, name string) Expression {
	return &namedExpression{
		relationName: relationName,
		name:         name,
	}
}

func (e namedExpression) ValueFrom(row shared.Row) (interface{}, error) {
	matches := 0
	for i, field := range row.Fields {
		if field.Name == e.name {
			if field.RelationName == e.relationName {
				return row.Values[i], nil
			}

			matches++
		}
	}

	if matches > 1 {
		return nil, fmt.Errorf("ambiguous field %s", e.name)
	}

	for i, field := range row.Fields {
		if field.Name == e.name {
			return row.Values[i], nil
		}
	}

	return nil, fmt.Errorf("unknown field %s", e.name)
}
