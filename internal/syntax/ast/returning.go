package ast

import (
	projection1 "github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/shared/impls"
)

func wrapReturning(node queries.Node, table impls.Table, alias string, expressions []projection1.ProjectionExpression) (queries.Node, error) {
	var aliasedTables []projection1.AliasedTable
	if alias != "" {
		aliasedTables = append(aliasedTables, projection1.AliasedTable{
			TableName: table.Name(),
			Alias:     alias,
		})
	}

	return projection.NewProjection(node, expressions, aliasedTables...)
}
