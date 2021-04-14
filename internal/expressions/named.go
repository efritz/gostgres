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

func (e namedExpression) Name() string {
	return e.name
}

func (e namedExpression) ValueFrom(row shared.Row) (interface{}, error) {
	index, err := e.indexFor(row.Fields)
	if err != nil {
		return nil, err
	}

	return row.Values[index], nil
}

func (e namedExpression) indexFor(fields []shared.Field) (int, error) {
	var unqualifiedIndexes []int
	for index, field := range fields {
		if field.Name != e.name {
			continue
		}

		if field.RelationName == e.relationName {
			return index, nil
		}

		unqualifiedIndexes = append(unqualifiedIndexes, index)
	}

	if e.relationName == "" {
		if len(unqualifiedIndexes) == 1 {
			return unqualifiedIndexes[0], nil
		}
		if len(unqualifiedIndexes) > 1 {
			return 0, fmt.Errorf("ambiguous field %s", e.name)
		}
	}

	return 0, fmt.Errorf("unknown field %s", e.name)
}
