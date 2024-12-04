package mutation

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	projectionHelpers "github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/ast"
)

type TargetTable struct {
	Name      string
	AliasName string
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

func joinNodes(left plan.LogicalNode, expressions []*ast.TableExpression) plan.LogicalNode {
	for _, expression := range expressions {
		right, err := expression.Build()
		if err != nil {
			return nil
		}

		left = plan.NewJoin(left, right, nil)
	}

	return left
}
