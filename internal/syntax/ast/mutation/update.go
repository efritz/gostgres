package mutation

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/projection"
	mutationNodes "github.com/efritz/gostgres/internal/execution/queries/nodes/mutation"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/execution/queries/plan/mutation"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/ast"
)

type UpdateBuilder struct {
	Target    TargetTable
	Updates   []SetExpression
	From      []*ast.TableExpression
	Where     impls.Expression
	Returning []projection.ProjectionExpression

	table     impls.Table
	returning *projection.Projection
}

type SetExpression struct {
	Name       string
	Expression impls.Expression
}

func (b *UpdateBuilder) Resolve(ctx *impls.NodeResolutionContext) error {
	table, ok := ctx.Catalog().Tables.Get(b.Target.Name)
	if !ok {
		return fmt.Errorf("unknown table %q", b.Target.Name)
	}
	b.table = table

	// TODO - resolve set expressions

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

	for _, from := range b.From {
		if err := from.Resolve(ctx); err != nil {
			return err
		}

		joinFields := from.TableFields()
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

func (b *UpdateBuilder) Build() (plan.LogicalNode, error) {
	node := plan.NewAccess(b.table)

	aliased, err := aliasTableNameForMutataion(node, b.Target.AliasName)
	if err != nil {
		return nil, err
	}
	node = aliased

	if b.From != nil {
		node = joinNodes(node, b.From)
	}

	setExpressions := make([]mutationNodes.SetExpression, len(b.Updates))
	for i, setExpression := range b.Updates {
		setExpressions[i] = mutationNodes.SetExpression{
			Name:       setExpression.Name,
			Expression: setExpression.Expression,
		}
	}

	return mutation.NewUpdate(node, b.table, b.Target.AliasName, setExpressions, b.Where, b.returning)
}
