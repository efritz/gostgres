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

type UpdateBuilder struct {
	TableDescription     TableDescription
	SetExpressions       []SetExpression
	FromExpressions      []TableExpressionDescription
	WhereExpression      impls.Expression
	ReturningExpressions []projection.ProjectionExpression
}

type SetExpression struct {
	Name       string
	Expression impls.Expression
}

func (b *UpdateBuilder) Build(ctx BuildContext) (queries.Node, error) {
	table, ok := ctx.Tables.Get(b.TableDescription.Name)
	if !ok {
		return nil, fmt.Errorf("unknown table %s", b.TableDescription.Name)
	}

	node := access.NewAccess(table)
	if b.TableDescription.AliasName != "" {
		node = alias.NewAlias(node, b.TableDescription.AliasName)
	}

	if b.FromExpressions != nil {
		node = joinNodes(ctx, node, b.FromExpressions)
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
		projection.NewTableWildcardProjectionExpression(relationName),
	})
	if err != nil {
		return nil, err
	}

	node = alias.NewAlias(node, b.TableDescription.Name)

	setExpressions := make([]mutation.SetExpression, len(b.SetExpressions))
	for i, setExpression := range b.SetExpressions {
		setExpressions[i] = mutation.SetExpression{
			Name:       setExpression.Name,
			Expression: setExpression.Expression,
		}
	}

	return mutation.NewUpdate(node, table, setExpressions, b.TableDescription.AliasName, b.ReturningExpressions)
}
