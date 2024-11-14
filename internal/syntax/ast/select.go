package ast

import (
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	projectionHelpers "github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/aggregate"
	"github.com/efritz/gostgres/internal/execution/queries/combination"
	"github.com/efritz/gostgres/internal/execution/queries/filter"
	"github.com/efritz/gostgres/internal/execution/queries/limit"
	"github.com/efritz/gostgres/internal/execution/queries/order"
	projection "github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/types"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

type SelectBuilder struct {
	Select *SimpleSelectDescription
	Order  impls.OrderExpression
	Limit  *int
	Offset *int

	fields []fields.Field
}

type SimpleSelectDescription struct {
	SelectExpressions []projectionHelpers.ProjectionExpression
	From              *TableExpression
	Where             impls.Expression
	Groupings         []impls.Expression
	Combinations      []*CombinationDescription
}

type CombinationDescription struct {
	Type     tokens.TokenType
	Distinct bool
	Select   TableReferenceOrExpression
}

func (b *SelectBuilder) Resolve(ctx impls.NodeResolutionContext) error {
	if err := b.Select.From.Resolve(ctx); err != nil {
		return err
	}

	fromFields := b.Select.From.TableFields()

	for _, c := range b.Select.Combinations {
		if err := c.Select.Resolve(ctx); err != nil {
			return err
		}
	}

	projection, err := projectionHelpers.NewProjection("", fromFields, b.Select.SelectExpressions)
	if err != nil {
		return err
	}

	b.fields = projection.Fields()
	return nil
}

func (b *SelectBuilder) TableFields() []fields.Field {
	return slices.Clone(b.fields)
}

func (b *SelectBuilder) Build() (queries.Node, error) {
	node, err := b.Select.From.Build()
	if err != nil {
		return nil, err
	}

	if b.Select.Where != nil {
		node = filter.NewFilter(node, b.Select.Where)
	}

	if b.Select.Groupings != nil {
		projectedExpressions, err := projectionHelpers.ExpandProjection(node.Fields(), b.Select.SelectExpressions)
		if err != nil {
			return nil, err
		}

	selectLoop:
		for _, selectExpression := range projectedExpressions {
			if len(expressions.Fields(selectExpression.Expression)) > 0 {
				alias := expressions.NewNamed(fields.NewField("", selectExpression.Alias, types.TypeAny, fields.NonInternalField))

				for _, grouping := range b.Select.Groupings {
					if grouping.Equal(selectExpression.Expression) || grouping.Equal(alias) {
						continue selectLoop
					}
				}

				// TODO - more lenient validation
				// return nil,  fmt.Errorf("%q not in group by", expression)
			}
		}

		node = aggregate.NewHashAggregate(node, b.Select.Groupings, b.Select.SelectExpressions)
	}

	if len(b.Select.Combinations) > 0 {
		if b.Select.Groupings == nil {
			node, err = projection.NewProjection(node, b.Select.SelectExpressions)
			if err != nil {
				return nil, err
			}
		}

		for _, c := range b.Select.Combinations {
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

	if b.Order != nil {
		node = order.NewOrder(node, b.Order)
	}
	if b.Offset != nil {
		node = limit.NewOffset(node, *b.Offset)
	}
	if b.Limit != nil {
		node = limit.NewLimit(node, *b.Limit)
	}

	if b.Select.Groupings == nil && len(b.Select.Combinations) == 0 {
		node, err = projection.NewProjection(node, b.Select.SelectExpressions)
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}
