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
	Target    TargetTable
	Updates   []SetExpression
	From      []TableExpression
	Where     impls.Expression
	Returning []projection.ProjectionExpression
}

type SetExpression struct {
	Name       string
	Expression impls.Expression
}

func (b *UpdateBuilder) Build(ctx BuildContext) (queries.Node, error) {
	table, ok := ctx.Tables.Get(b.Target.Name)
	if !ok {
		return nil, fmt.Errorf("unknown table %s", b.Target.Name)
	}

	node := access.NewAccess(table)
	if b.Target.AliasName != "" {
		node = alias.NewAlias(node, b.Target.AliasName)
	}

	if b.From != nil {
		node = joinNodes(ctx, node, b.From)
	}

	if b.Where != nil {
		node = filter.NewFilter(node, b.Where)
	}

	relationName := b.Target.Name
	if b.Target.AliasName != "" {
		relationName = b.Target.AliasName
	}
	tidField := fields.NewField(relationName, rows.TIDName, types.TypeBigInteger)

	node, err := projection.NewProjection(node, []projection.ProjectionExpression{
		projection.NewAliasProjectionExpression(expressions.NewNamed(tidField), rows.TIDName),
		projection.NewTableWildcardProjectionExpression(relationName),
	})
	if err != nil {
		return nil, err
	}

	node = alias.NewAlias(node, b.Target.Name)

	setExpressions := make([]mutation.SetExpression, len(b.Updates))
	for i, setExpression := range b.Updates {
		setExpressions[i] = mutation.SetExpression{
			Name:       setExpression.Name,
			Expression: setExpression.Expression,
		}
	}

	return mutation.NewUpdate(node, table, setExpressions, b.Target.AliasName, b.Returning)
}