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

	table     impls.Table
	returning *projection.Projection
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

	if b.Target.AliasName != "" {
		for i, field := range baseFields {
			baseFields[i] = field.WithRelationName(b.Target.AliasName)
		}
	}

	ctx.PushScope()
	defer ctx.PopScope()

	ctx.Bind(baseFields...)

	for _, e := range b.Using {
		if err := e.Resolve(ctx); err != nil {
			return err
		}

		joinFields := e.TableFields()
		ctx.Bind(joinFields...)
		baseFields = append(baseFields, joinFields...)
	}

	resolved, err := ast.ResolveExpression(ctx, b.Where, nil, false)
	if err != nil {
		return err
	}
	b.Where = resolved

	// TODO - resolve projection
	returning, err := returningProjection(b.table, b.Target.AliasName, b.Returning)
	if err != nil {
		return err
	}
	b.returning = returning

	return nil
}

func (b *DeleteBuilder) Build() (plan.LogicalNode, error) {
	node := plan.NewAccess(b.table)

	aliased, err := aliasTableNameForMutataion(node, b.Target.AliasName)
	if err != nil {
		return nil, err
	}
	node = aliased

	if len(b.Using) > 0 {
		node = joinNodes(node, b.Using)
	}

	return mutation.NewDelete(node, b.table, b.Target.AliasName, b.Where, b.returning)
}
