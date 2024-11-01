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

func (b *ValuesBuilder) Resolve(ctx *context.ResolveContext) error {
	return nil
}

func (ValuesBuilder) tableExpression() {}

func (b *ValuesBuilder) Build() (queries.Node, error) {
	return access.NewValues(b.Fields, b.Expressions), nil
}
