package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/access"
	"github.com/efritz/gostgres/internal/execution/queries/alias"
	"github.com/efritz/gostgres/internal/execution/queries/filter"
	"github.com/efritz/gostgres/internal/execution/queries/mutation"
	"github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

type DeleteBuilder struct {
	TableDescription     TableDescription
	UsingExpressions     []TableExpressionDescription
	WhereExpression      impls.Expression
	ReturningExpressions []projection.ProjectionExpression
}

func (b *DeleteBuilder) Build(ctx BuildContext) (queries.Node, error) {
	table, ok := ctx.Tables.Get(b.TableDescription.Name)
	if !ok {
		return nil, fmt.Errorf("unknown table %s", b.TableDescription.Name)
	}

	node := access.NewAccess(table)
	if b.TableDescription.AliasName != "" {
		node = alias.NewAlias(node, b.TableDescription.AliasName)
	}
	if len(b.UsingExpressions) > 0 {
		node = joinNodes(ctx, node, b.UsingExpressions)
	}
	if b.WhereExpression != nil {
		node = filter.NewFilter(node, b.WhereExpression)
	}

	relationName := b.TableDescription.Name
	if b.TableDescription.AliasName != "" {
		relationName = b.TableDescription.AliasName
	}
	tidField := fields.NewField(relationName, rows.TIDName, types.TypeBigInteger)

	node, err := projection.NewProjection(node, []projection.ProjectionExpression{
		projection.NewAliasProjectionExpression(expressions.NewNamed(tidField), rows.TIDName),
	})
	if err != nil {
		return nil, err
	}

	return mutation.NewDelete(node, table, b.TableDescription.AliasName, b.ReturningExpressions)
}
