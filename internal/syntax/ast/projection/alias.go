package projection

import (
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/ast/context"
)

type AliasProjection struct {
	Expression impls.Expression
	Alias      string
}

func NewAliasProjection(expression impls.Expression, alias string) *AliasProjection {
	return &AliasProjection{
		Expression: expression,
		Alias:      alias,
	}
}

func (p *AliasProjection) Expand(ctx *context.ResolverContext) ([]AliasedExpression, error) {
	return []AliasedExpression{{Expression: p.Expression, Alias: p.Alias}}, nil
}
