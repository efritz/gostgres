package mutation

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projection"
	concreteJoin "github.com/efritz/gostgres/internal/execution/queries/nodes/join"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	logicalJoin "github.com/efritz/gostgres/internal/execution/queries/plan/join"
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
	ctx.Bind(tableFields)

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
	ctx.Bind(fields)

	return ast.ResolveProjection(ctx, "", fields, expressions, aliasedTables)
}

// TODO - share with table expressions
func joinNodes(node plan.LogicalNode, expressions []*ast.TableExpression) (plan.LogicalNode, error) {
	if len(expressions) == 0 {
		return node, nil
	}

	joinNode := logicalJoin.NewJoinLeafNode(node)

	for _, e := range expressions {
		right, err := e.Build()
		if err != nil {
			return nil, err
		}

		joinNode = logicalJoin.NewJoinInternalNode(joinNode, logicalJoin.NewJoinLeafNode(right), logicalJoin.JoinOperator{
			JoinType:  concreteJoin.JoinTypeInner, // TODO - cross join?
			Condition: nil,
		})
	}

	return joinNode, nil
}
