package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projector"
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
	"github.com/efritz/gostgres/internal/syntax/ast/context"
)

type UpdateBuilder struct {
	Target    TargetTable
	Updates   []SetExpression
	From      []*TableExpression
	Where     impls.Expression
	Returning []projector.ProjectionExpression

	table impls.Table
}

type SetExpression struct {
	Name       string
	Expression impls.Expression
}

func (b *UpdateBuilder) Resolve(ctx *context.ResolveContext) error {
	table, ok := ctx.Tables.Get(b.Target.Name)
	if !ok {
		return fmt.Errorf("unknown table %q", b.Target.Name)
	}
	b.table = table

	for _, from := range b.From {
		if err := from.Resolve(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (b *UpdateBuilder) Build() (queries.Node, error) {
	node := access.NewAccess(b.table)
	if b.Target.AliasName != "" {
		node = alias.NewAlias(node, b.Target.AliasName)
	}

	if b.From != nil {
		node = joinNodes(node, b.From)
	}

	if b.Where != nil {
		node = filter.NewFilter(node, b.Where)
	}

	relationName := b.Target.Name
	if b.Target.AliasName != "" {
		relationName = b.Target.AliasName
	}
	tidField := fields.NewField(relationName, rows.TIDName, types.TypeBigInteger)

	node, err := projection.NewProjection(node, []projector.ProjectionExpression{
		projector.NewAliasProjectionExpression(expressions.NewNamed(tidField), rows.TIDName),
		projector.NewTableWildcardProjectionExpression(relationName),
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

	return mutation.NewUpdate(node, b.table, setExpressions, b.Target.AliasName, b.Returning)
}
