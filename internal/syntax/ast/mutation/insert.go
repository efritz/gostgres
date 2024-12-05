package mutation

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/execution/queries/plan/mutation"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/ast"
)

type InsertBuilder struct {
	Target      TargetTable
	ColumnNames []string
	Source      ast.TableReferenceOrExpression
	Returning   []projection.ProjectionExpression

	table     impls.Table
	returning *projection.Projection
}

func (b *InsertBuilder) Resolve(ctx *impls.NodeResolutionContext) error {
	table, ok := ctx.Catalog().Tables.Get(b.Target.Name)
	if !ok {
		return fmt.Errorf("unknown table %q", b.Target.Name)
	}
	b.table = table

	// TODO - resolve column names

	if err := b.Source.Resolve(ctx); err != nil {
		return err
	}

	returning, err := resolveReturning(ctx, b.table, b.Target.AliasName, b.Returning)
	if err != nil {
		return err
	}
	b.returning = returning

	return nil
}

func (b *InsertBuilder) Build() (plan.LogicalNode, error) {
	node, err := b.Source.Build()
	if err != nil {
		return nil, err
	}

	return mutation.NewInsert(node, b.table, b.ColumnNames, b.returning)
}
