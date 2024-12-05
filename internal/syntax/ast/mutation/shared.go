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

func resolveMutationProjection(
	ctx *impls.NodeResolutionContext,
	tableName string,
	aliasName string,
	tableFields []fields.Field,
) (*projection.Projection, error) {
	name := aliasName
	if name == "" {
		name = tableName
	}

	projectionExpressions := []projection.ProjectionExpression{
		projection.NewAliasedExpression(expressions.NewNamed(fields.TIDField), "$tid", true),
		projection.NewWildcardProjectionExpression(),
	}

	ctx.PushScope()
	defer ctx.PopScope()
	ctx.Bind(tableFields...)

	return ast.ResolveProjection(ctx, name, tableFields, projectionExpressions, nil)
}

func resolveReturning(
	ctx *impls.NodeResolutionContext,
	table impls.Table,
	alias string,
	expressions []projection.ProjectionExpression,
) (*projection.Projection, error) {
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

	ctx.PushScope()
	defer ctx.PopScope()
	ctx.Bind(fields...)

	return ast.ResolveProjection(ctx, "", fields, expressions, aliasedTables)
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
