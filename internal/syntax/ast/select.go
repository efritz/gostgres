package ast

import (
	"fmt"
	"strings"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projector"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/aggregate"
	"github.com/efritz/gostgres/internal/execution/queries/combination"
	"github.com/efritz/gostgres/internal/execution/queries/filter"
	"github.com/efritz/gostgres/internal/execution/queries/limit"
	"github.com/efritz/gostgres/internal/execution/queries/order"
	"github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/types"
	"github.com/efritz/gostgres/internal/syntax/ast/context"
	astprojection "github.com/efritz/gostgres/internal/syntax/ast/projection"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

type SelectBuilder struct {
	Select *SimpleSelectDescription
	Order  impls.OrderExpression
	Limit  *int
	Offset *int
}

func (s *SelectBuilder) String() string {
	return fmt.Sprintf("SELECT %v\nORDER %v\nLIMIT %v\nOFFSET %v\n", s.Select, s.Order, s.Limit, s.Offset)
}

func (b *SelectBuilder) Resolve(ctx *context.ResolverContext) ([]fields.Field, error) {
	return b.ResolveWithAlias(ctx, nil)
}

func (b *SelectBuilder) ResolveWithAlias(ctx *context.ResolverContext, alias *TableAlias) ([]fields.Field, error) {
	fields, err := b.Select.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	if alias != nil {
		panic("OH NO") // TODO
	}

	// TODO
	return fields, nil
}

func (b *SelectBuilder) Build() (queries.Node, error) {
	node, err := b.Select.Build()
	if err != nil {
		return nil, err
	}

	if b.Order != nil {
		node = order.NewOrder(node, b.Order)
	}
	if b.Offset != nil {
		node = limit.NewOffset(node, *b.Offset)
	}
	if b.Limit != nil {
		node = limit.NewLimit(node, *b.Limit)
	}

	// TODO - must come last? Can order come first?
	if b.Select.SelectExpressions != nil {
		var e []projector.ProjectionExpression
		for _, p := range b.Select.SelectExpressions {
			_ = p // TODO
		}

		return projection.NewProjection(node, e)
	}
	return node, nil
}

//
//

type SimpleSelectDescription struct {
	SelectExpressions   []astprojection.Projection
	ResolvedExpressions []impls.Expression
	From                TableExpression
	Where               impls.Expression
	Groupings           []impls.Expression
	Combinations        []*CombinationDescription
}

func (ssd SimpleSelectDescription) String() string {
	var x []string
	for _, v := range ssd.ResolvedExpressions {
		x = append(x, fmt.Sprintf("%s", v))
	}

	return fmt.Sprintf("SELECT %s\nFROM %v\n\n", strings.Join(x, ", "), ssd.From)
}

type CombinationDescription struct {
	Type     tokens.TokenType
	Distinct bool
	Select   TableReferenceOrExpression
}

func (b *SimpleSelectDescription) Resolve(ctx *context.ResolverContext) ([]fields.Field, error) {
	ctx.SymbolTable.PushScope()
	defer ctx.SymbolTable.PopScope()

	fs1, err := b.From.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	// where, err := mapExpression(ctx, b.Where)
	// if err != nil {
	// 	return nil, err
	// }
	// b.Where = where
	// fmt.Printf("> %#v\n", where)
	// _ = where // TODO

	for _, t := range b.Combinations {
		fs2, err := t.Select.Resolve(ctx)
		if err != nil {
			return nil, err
		}

		_, _ = fs1, fs2 // TODO - compare types
	}

	var projectedFields []fields.Field
	for _, selectExpression := range b.SelectExpressions {
		exprs, err := selectExpression.Expand(ctx)
		fmt.Printf("> SELECT: %#v\n", selectExpression)
		for _, expr := range exprs {
			fmt.Printf("EXPRS: %#v\n", expr)
		}
		fmt.Printf("\n\n")
		if err != nil {
			return nil, err
		}

		for _, e := range exprs {
			name := e.Alias
			if name == "" {
				name = "?column?"
			}

			var typ types.Type = types.TypeAny

			fmt.Printf("TRYING TO RESOLVE %#v\n", e.Expression)
			if expr, err := mapExpression(ctx, e.Expression); err != nil {
				fmt.Printf("Error %#v: %s\n", e.Expression, err)
			} else {
				b.ResolvedExpressions = append(b.ResolvedExpressions, expr)
				fmt.Printf("Mapped %#v: %#v\n", e.Expression, expr)
			}

			projectedFields = append(projectedFields, fields.NewField("", name, typ))
		}
	}

	return projectedFields, nil
}

func (b *SimpleSelectDescription) Build() (queries.Node, error) {
	node, err := b.From.Build()
	if err != nil {
		return nil, err
	}

	if b.Where != nil {
		node = filter.NewFilter(node, b.Where)
	}

	var selectExpressions []projector.ProjectionExpression
	for _, p := range b.SelectExpressions {
		_ = p // TODOk
	}

	if b.Groupings != nil {
	selectLoop:
		for _, selectExpression := range selectExpressions {
			expression, alias, ok := projector.UnwrapAlias(selectExpression)
			if !ok {
				return nil, fmt.Errorf("cannot unwrap alias %q", selectExpression)
			}

			if len(expressions.Fields(expression)) > 0 {
				for _, grouping := range b.Groupings {
					if grouping.Equal(expression) || grouping.Equal(expressions.NewNamed(fields.NewField("", alias, types.TypeAny))) {
						continue selectLoop
					}
				}

				// TODO - more lenient validation
				// return nil,  fmt.Errorf("%q not in group by", expression)
			}
		}

		node = aggregate.NewHashAggregate(node, b.Groupings, selectExpressions)
		selectExpressions = nil
		b.SelectExpressions = nil
	}

	if len(b.Combinations) != 0 {
		if selectExpressions != nil {
			newNode, err := projection.NewProjection(node, selectExpressions)
			if err != nil {
				return nil, err
			}
			node = newNode
			selectExpressions = nil
			b.SelectExpressions = nil
		}

		for _, c := range b.Combinations {
			var factory func(left, right queries.Node, distinct bool) (queries.Node, error)
			switch c.Type {
			case tokens.TokenTypeUnion:
				factory = combination.NewUnion
			case tokens.TokenTypeIntersect:
				factory = combination.NewIntersect
			case tokens.TokenTypeExcept:
				factory = combination.NewExcept
			}

			right, err := c.Select.Build()
			if err != nil {
				return nil, err
			}

			newNode, err := factory(node, right, c.Distinct)
			if err != nil {
				return nil, err
			}
			node = newNode
		}
	}

	return node, nil
}
