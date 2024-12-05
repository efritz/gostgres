package mutation

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/execution/queries/plan/mutation"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/ast"
)

type DeleteBuilder struct {
	Target    TargetTable
	Using     []*ast.TableExpression
	Where     impls.Expression
	Returning []projection.ProjectionExpression

	table           impls.Table
	aliasProjection *projection.Projection
	returning       *projection.Projection
}

func (b *DeleteBuilder) Resolve(ctx *impls.NodeResolutionContext) error {
	table, ok := ctx.Catalog().Tables.Get(b.Target.Name)
	if !ok {
		return fmt.Errorf("unknown table %q", b.Target.Name)
	}
	b.table = table

	var baseFields []fields.Field
	for _, field := range table.Fields() {
		baseFields = append(baseFields, field.Field)
	}

	p, err := resolveMutationProjection(ctx, table.Name(), b.Target.AliasName, baseFields)
	if err != nil {
		return err
	}
	b.aliasProjection = p

	ctx.PushScope()
	defer ctx.PopScope()
	ctx.Bind(p.Fields()...)

	for _, e := range b.Using {
		if err := e.Resolve(ctx); err != nil {
			return err
		}

		ctx.Bind(e.TableFields()...)
	}

	resolved, err := ast.ResolveExpression(ctx, b.Where, nil, false)
	if err != nil {
		return err
	}
	b.Where = resolved

	returning, err := resolveReturning(ctx, b.table, b.Target.AliasName, b.Returning)
	if err != nil {
		return err
	}
	b.returning = returning

	return nil
}

func (b *DeleteBuilder) Build() (plan.LogicalNode, error) {
	node := plan.NewAccess(b.table)
	node = plan.NewProjection(node, b.aliasProjection)

	if len(b.Using) > 0 {
		node = joinNodes(node, b.Using)
	}

	return mutation.NewDelete(node, b.table, b.Target.AliasName, b.Where, b.returning)
}
