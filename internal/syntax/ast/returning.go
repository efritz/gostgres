package ast

import (
	projectionHelpers "github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/impls"
)

func wrapReturning(node nodes.LogicalNode, table impls.Table, alias string, expressions []projectionHelpers.ProjectionExpression) (nodes.LogicalNode, error) {
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

	return nodes.NewProjection(node, p), nil
}
