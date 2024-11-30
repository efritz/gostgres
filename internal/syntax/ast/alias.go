package ast

import (
	projectionHelpers "github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/opt"
)

func aliasTableName(node opt.LogicalNode, name string) (opt.LogicalNode, error) {
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

	return opt.NewProjection(node, p), nil
}
