package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/opt"
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
	table, ok := ctx.Catalog().Tables.Get(b.Target.Name)
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

func (b *DeleteBuilder) Build() (opt.LogicalNode, error) {
	node := opt.NewAccess(b.table)
	if b.Target.AliasName != "" {
		aliased, err := aliasTableName(node, b.Target.AliasName)
		if err != nil {
			return nil, err
		}
		node = aliased
	}
	if len(b.Using) > 0 {
		node = joinNodes(node, b.Using)
	}
	if b.Where != nil {
		node = opt.NewFilter(node, b.Where)
	}

	delete, err := opt.NewDelete(node, b.table, b.Target.AliasName)
	if err != nil {
		return nil, err
	}

	return wrapReturning(delete, b.table, b.Target.AliasName, b.Returning)
}
