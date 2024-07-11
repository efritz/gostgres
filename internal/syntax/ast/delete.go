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

type DeleteBuilder struct {
	Target    TargetTable
	Using     []TableExpression
	Where     impls.Expression
	Returning []projector.ProjectionExpression
}

func (b *DeleteBuilder) Resolve(ctx *context.ResolverContext) ([]fields.Field, error) {
	return nil, fmt.Errorf("delete resolve unimplemented")
}

func (b *DeleteBuilder) Build() (queries.Node, error) {
	// table, ok := ctx.Tables.Get(b.Target.Name)
	// if !ok {
	// 	return nil, fmt.Errorf("unknown table %q", b.Target.Name)
	// }
	var table impls.Table // TODO

	node := access.NewAccess(table)
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
	tidField := fields.NewField(relationName, rows.TIDName, types.TypeBigInteger)

	node, err := projection.NewProjection(node, []projector.ProjectionExpression{
		projector.NewAliasProjectionExpression(expressions.NewNamed(tidField), rows.TIDName),
	})
	if err != nil {
		return nil, err
	}

	return mutation.NewDelete(node, table, b.Target.AliasName, b.Returning)
}
