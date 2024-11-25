package ast

import (
	projectionHelpers "github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries"
	projection "github.com/efritz/gostgres/internal/execution/queries/projection"
)

func aliasTableName(node queries.Node, name string) (queries.Node, error) {
	p, err := projectionHelpers.NewProjectionFromProjectionExpressions(
		name,
		node.Fields(),
		[]projectionHelpers.ProjectionExpression{
			projectionHelpers.NewWildcardProjectionExpression(),
		},
	)
	if err != nil {
		return nil, err
	}

	return projection.NewProjection(node, p), nil
}
