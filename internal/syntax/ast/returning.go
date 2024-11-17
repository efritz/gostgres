package ast

import (
	projectionHelpers "github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries"
	projection "github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/shared/impls"
)

func wrapReturning(node queries.Node, table impls.Table, alias string, expressions []projectionHelpers.ProjectionExpression) (queries.Node, error) {
	var aliasedTables []projectionHelpers.AliasedTable
	if alias != "" {
		aliasedTables = append(aliasedTables, projectionHelpers.AliasedTable{
			TableName: table.Name(),
			Alias:     alias,
		})
	}

	p, err := projectionHelpers.NewProjectionFromProjectionExpressions(node.Name(), node.Fields(), expressions, aliasedTables...)
	if err != nil {
		return nil, err
	}

	return projection.NewProjection(node, p), nil
}
