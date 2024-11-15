package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/access"
	"github.com/efritz/gostgres/internal/execution/queries/alias"
	"github.com/efritz/gostgres/internal/execution/queries/filter"
	"github.com/efritz/gostgres/internal/execution/queries/mutation"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type DeleteBuilder struct {
	Target    TargetTable
	Using     []*TableExpression
	Where     impls.Expression
	Returning []projection.ProjectionExpression

	table impls.Table
}

func (b *DeleteBuilder) Resolve(ctx *impls.NodeResolutionContext) error {
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

	delete, err := mutation.NewDelete(node, b.table, b.Target.AliasName)
	if err != nil {
		return nil, err
	}

	return wrapReturning(delete, b.table, b.Target.AliasName, b.Returning)
}
