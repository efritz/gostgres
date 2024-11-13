package ast

import (
	"github.com/efritz/gostgres/internal/execution/projector"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

func wrapReturning(node queries.Node, table impls.Table, alias string, expressions []projector.ProjectionExpression) (queries.Node, error) {
	var fields []fields.Field
	for _, field := range table.Fields() {
		fields = append(fields, field.Field)
	}

	var aliasedTables []projector.AliasedTable
	if alias != "" {
		aliasedTables = append(aliasedTables, projector.AliasedTable{
			TableName: table.Name(),
			Alias:     alias,
		})
	}

	projectedExpressions, err := projector.ExpandProjection(fields, expressions, aliasedTables...)
	if err != nil {
		return nil, err
	}

	return projection.NewProjectionFromProjectedExpressions(node, projectedExpressions), nil
}
