package projection

import (
	"fmt"

	"github.com/efritz/gostgres/internal/syntax/ast/context"
)

type TableWildcardProjection struct {
	tableName string
}

func NewTableWildcardProjection(tableName string) *TableWildcardProjection {
	return &TableWildcardProjection{
		tableName: tableName,
	}
}

func (p *TableWildcardProjection) Expand(ctx *context.ResolverContext) ([]AliasedExpression, error) {
	for i := 0; i < len(ctx.SymbolTable.Scopes); i++ {
		s := ctx.SymbolTable.Scopes[len(ctx.SymbolTable.Scopes)-i-1]

		for tableName, tableDescription := range s.Symbols {
			if tableName != p.tableName {
				continue
			}

			var exprs []AliasedExpression
			for i, field := range tableDescription.Fields {
				exprs = append(exprs, AliasedExpression{
					Expression: tableDescription.Expressions[i],
					Alias:      field.Name(),
				})
			}

			return exprs, nil
		}
	}

	return nil, fmt.Errorf("unknown table %q", p.tableName)
}
