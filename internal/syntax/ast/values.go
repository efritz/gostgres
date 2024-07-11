package ast

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/access"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/ast/context"
)

type ValuesBuilder struct {
	Fields      []fields.Field
	Expressions [][]impls.Expression
}

func (b *ValuesBuilder) Resolve(ctx *context.ResolverContext) ([]fields.Field, error) {
	return b.ResolveWithAlias(ctx, nil)
}

func (b *ValuesBuilder) ResolveWithAlias(ctx *context.ResolverContext, alias *TableAlias) ([]fields.Field, error) {
	if alias != nil {
		panic("OH NO") // TODO
	}

	return b.Fields, nil // TODO - check expression types
}

func (b *ValuesBuilder) Build() (queries.Node, error) {
	return access.NewValues(b.Fields, b.Expressions), nil
}
