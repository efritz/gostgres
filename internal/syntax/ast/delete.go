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
	"github.com/efritz/gostgres/internal/shared/types"
	"github.com/efritz/gostgres/internal/syntax/ast/context"
)

type DeleteBuilder struct {
	Target    TargetTable
	Using     []*TableExpression
	Where     impls.Expression
	Returning []projector.ProjectionExpression

	table impls.Table
}

func (b *DeleteBuilder) Resolve(ctx *context.ResolveContext) error {
	table, ok := ctx.Catalog.Tables.Get(b.Target.Name)
	if !ok {
		return fmt.Errorf("unknown table %q", b.Target.Name)
	}
	b.table = table

	for _, e := range b.Using {
		if err := e.Resolve(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (b *DeleteBuilder) Build() (queries.Node, error) {
	node := access.NewAccess(b.table)
	if b.Target.AliasName != "" {
		node = alias.NewAlias(node, b.Target.AliasName)
	}
	if len(b.Using) > 0 {
		node = joinNodes(node, b.Using)
	}
	if b.Where != nil {
		node = filter.NewFilter(node, b.Where)
	}

	relationName := b.Target.Name
	if b.Target.AliasName != "" {
		relationName = b.Target.AliasName
	}
	tidField := fields.NewField(relationName, "tid", types.TypeBigInteger, fields.InternalFieldTid)

	node, err := projection.NewProjection(node, []projector.ProjectionExpression{
		projector.NewAliasProjectionExpression(expressions.NewNamed(tidField), "tid"),
	})
	if err != nil {
		return nil, err
	}

	return mutation.NewDelete(node, b.table, b.Target.AliasName, b.Returning)
}
