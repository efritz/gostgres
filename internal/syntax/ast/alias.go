package ast

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	projectionHelpers "github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/shared/fields"
)

func aliasTableName(node plan.LogicalNode, name string) (plan.LogicalNode, error) {
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

	return plan.NewProjection(node, p), nil
}

func aliasTableNameForMutataion(node plan.LogicalNode, name string) (plan.LogicalNode, error) {
	if name == "" {
		name = node.Name()
	}

	p, err := projectionHelpers.NewProjectionFromProjectionExpressions(
		name,
		node.Fields(),
		[]projectionHelpers.ProjectionExpression{
			projectionHelpers.NewAliasedExpression(expressions.NewNamed(fields.TIDField), "$tid", true),
			projectionHelpers.NewWildcardProjectionExpression(),
		},
	)
	if err != nil {
		return nil, err
	}

	return plan.NewProjection(node, p), nil
}
