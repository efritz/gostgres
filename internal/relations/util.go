package relations

import (
	"strings"

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
	for i := range fields {
		fields[i].RelationName = relationName
	}

	return fields
}

func indent(level int) string {
	return strings.Repeat(" ", level*4)
}
