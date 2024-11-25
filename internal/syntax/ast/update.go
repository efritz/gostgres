package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/access"
	"github.com/efritz/gostgres/internal/execution/queries/filter"
	"github.com/efritz/gostgres/internal/execution/queries/mutation"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type UpdateBuilder struct {
	Target    TargetTable
	Updates   []SetExpression
	From      []*TableExpression
	Where     impls.Expression
	Returning []projection.ProjectionExpression

	table impls.Table
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
		aliased, err := aliasTableName(node, b.Target.AliasName)
		if err != nil {
			return nil, err
		}
		node = aliased
	}

	if b.From != nil {
		node = joinNodes(node, b.From)
	}

	if b.Where != nil {
		node = filter.NewFilter(node, b.Where)
	}

	setExpressions := make([]mutation.SetExpression, len(b.Updates))
	for i, setExpression := range b.Updates {
		setExpressions[i] = mutation.SetExpression{
			Name:       setExpression.Name,
			Expression: setExpression.Expression,
		}
	}

	update, err := mutation.NewUpdate(node, b.table, b.Target.AliasName, setExpressions)
	if err != nil {
		return nil, err
	}

	return wrapReturning(update, b.table, b.Target.AliasName, b.Returning)
}
