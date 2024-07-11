package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/ast/context"
)

type Resolver interface {
	Resolve(ctx *context.ResolverContext) ([]fields.Field, error)
}

func mapExpression(ctx *context.ResolverContext, e impls.Expression) (impls.Expression, error) {
	if e == nil {
		return nil, nil
	}

	return e.Map(func(e impls.Expression) (impls.Expression, error) {
		if named, ok := e.(expressions.NamedExpression); ok {
			fmt.Printf("NAMED: %#v\n", named)
			f := named.Field()
			relationName := f.RelationName()
			name := f.Name()
			if relationName == "" {
				// TODO - need to worry about aliasing?
				var matches []fields.Field
				for i := 0; i < len(ctx.SymbolTable.Scopes); i++ {
					s := ctx.SymbolTable.Scopes[len(ctx.SymbolTable.Scopes)-i-1]

					for _, tableDescription := range s.Symbols {
						for _, field := range tableDescription.Fields {
							if field.Name() == name {
								matches = append(matches, field)
							}
						}
					}
				}

				if len(matches) == 0 {
					return nil, fmt.Errorf("unknown field %q", name)
				} else if len(matches) > 1 {
					return nil, fmt.Errorf("ambiguous field %q", name)
				} else {
					return expressions.NewNamed(matches[0]), nil
				}
			} else {
				for i := 0; i < len(ctx.SymbolTable.Scopes); i++ {
					s := ctx.SymbolTable.Scopes[len(ctx.SymbolTable.Scopes)-i-1]

					tableDescription, ok := s.Symbols[relationName]
					if !ok {
						continue
					}

					for i, field := range tableDescription.Fields {
						if field.Name() == name {
							// TODO: .WithRelationName(tableDescription.UniqueName)
							return tableDescription.Expressions[i], nil // expressions.NewNamed(field), nil
						}
					}

					return nil, fmt.Errorf("unknown field %q in relation %q", name, relationName)
				}

				for _, s := range ctx.SymbolTable.Scopes {
					for k, v := range s.Symbols {
						fmt.Printf("> %s: %#v\n", k, v)
					}
				}
				fmt.Printf("[expr resolver] unknown relation %q\n", relationName)
				panic("NOPE")
				return nil, fmt.Errorf("[expr resolver] unknown relation %q", relationName)
			}
		}

		return e, nil
	})
}
