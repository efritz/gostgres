package projection

import (
	"github.com/efritz/gostgres/internal/syntax/ast/context"
)

type WildcardProjection struct{}

func NewWildcardProjection() *WildcardProjection {
	return &WildcardProjection{}
}

func (p *WildcardProjection) Expand(ctx *context.ResolverContext) ([]AliasedExpression, error) {
	var exprs []AliasedExpression
	for _, symbol := range ctx.SymbolTable.CurrentScope().Symbols {
		for i, field := range symbol.Fields {
			exprs = append(exprs, AliasedExpression{
				Expression: symbol.Expressions[i],
				Alias:      field.Name(),
			})
		}
	}

	return exprs, nil
}
