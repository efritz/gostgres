package nodes

import (
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

func copyFields(fields []shared.Field) []shared.Field {
	c := make([]shared.Field, len(fields))
	copy(c, fields)
	return c
}

func copyValues(values []interface{}) []interface{} {
	c := make([]interface{}, len(values))
	copy(c, values)
	return c
}

func updateRelationName(fields []shared.Field, relationName string) []shared.Field {
	fields = copyFields(fields)
	for i := range fields {
		fields[i].RelationName = relationName
	}

	return fields
}

func indent(level int) string {
	return strings.Repeat(" ", level*4)
}

func combineConjunctions(conjunctions []expressions.Expression) expressions.Expression {
	if len(conjunctions) == 0 {
		return nil
	}

	expression := conjunctions[0]
	for _, conjunction := range conjunctions[1:] {
		expression = expressions.NewAnd(expression, conjunction)
	}

	return expression
}
