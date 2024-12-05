package mutation

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projection"
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

	// TODO - share with resolveTableAlias
	projectionExpressions := []projection.ProjectionExpression{
		projection.NewAliasedExpression(expressions.NewNamed(fields.TIDField), "$tid", true),
		projection.NewWildcardProjectionExpression(),
	}
	projectedExpressions, err := projection.ExpandProjection(node.Fields(), projectionExpressions, nil)
	if err != nil {
		return nil, err
	}
	p, err := projection.NewProjection(name, projectedExpressions)
	if err != nil {
		return nil, err
	}

	return plan.NewProjection(node, p), nil
}

func returningProjection(table impls.Table, alias string, expressions []projection.ProjectionExpression) (*projection.Projection, error) {
	var fields []fields.Field
	for _, field := range table.Fields() {
		fields = append(fields, field.Field)
	}

	var aliasedTables []projection.AliasedTable
	if alias != "" {
		aliasedTables = append(aliasedTables, projection.AliasedTable{
			TableName: table.Name(),
			Alias:     alias,
		})
	}
	projectedExpressions, err := projection.ExpandProjection(fields, expressions, aliasedTables)
	if err != nil {
		return nil, err
	}
	return projection.NewProjection("", projectedExpressions)
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
