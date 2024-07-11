package projection

import (
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/ast/context"
)

type Projection interface {
	Expand(ctx *context.ResolverContext) ([]AliasedExpression, error)
}

type AliasedExpression struct {
	Expression impls.Expression
	Alias      string
}
