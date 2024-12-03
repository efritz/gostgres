package ast

import (
	projectionHelpers "github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

func returningProjection(table impls.Table, alias string, expressions []projectionHelpers.ProjectionExpression) (*projectionHelpers.Projection, error) {
	if len(expressions) == 0 {
		return projectionHelpers.NewProjectionFromProjectedExpressions("", nil)
	}

	var fields []fields.Field
	for _, field := range table.Fields() {
		fields = append(fields, field.Field)
	}

	var aliasedTables []projectionHelpers.AliasedTable
	if alias != "" {
		aliasedTables = append(aliasedTables, projectionHelpers.AliasedTable{
			TableName: table.Name(),
			Alias:     alias,
		})
	}

	return projectionHelpers.NewProjectionFromProjectionExpressions("", fields, expressions, aliasedTables...)
}
