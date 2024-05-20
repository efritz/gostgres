package expressions

import (
	"github.com/efritz/gostgres/internal/aggregates"
	"github.com/efritz/gostgres/internal/shared"
)

type AggregateExpression interface {
	Step(ctx ExpressionContext, row shared.Row) error
	Done(ctx ExpressionContext) (any, error)
}

func AsAggregate(ctx ExpressionContext, e Expression) AggregateExpression {
	var (
		results        []*constantExpression
		subExpressions []AggregateExpression
	)
	outerExpression := e.Map(func(e Expression) Expression {
		f, ok := e.(functionExpression)
		if !ok {
			return e
		}

		aggregate, ok := ctx.GetAggregate(f.name)
		if !ok {
			return e
		}

		placeholder := &constantExpression{}
		results = append(results, placeholder)
		subExpressions = append(subExpressions, &aggregateSubExpression{aggregate: aggregate, args: f.args})
		return placeholder
	})

	if len(subExpressions) > 0 {
		return &explodedAggregateExpression{
			results:         results,
			subExpressions:  subExpressions,
			outerExpression: outerExpression,
		}
	}

	return &nonAggregateExpression{
		expression: outerExpression,
	}
}

type explodedAggregateExpression struct {
	results         []*constantExpression
	subExpressions  []AggregateExpression
	outerExpression Expression
}

var _ AggregateExpression = &explodedAggregateExpression{}

func (e *explodedAggregateExpression) Step(ctx ExpressionContext, row shared.Row) error {
	for _, subexpression := range e.subExpressions {
		if err := subexpression.Step(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func (e *explodedAggregateExpression) Done(ctx ExpressionContext) (any, error) {
	for i, subExpression := range e.subExpressions {
		value, err := subExpression.Done(ctx)
		if err != nil {
			return nil, err
		}

		e.results[i].value = value
	}

	return e.outerExpression.ValueFrom(ctx, shared.Row{})
}

type aggregateSubExpression struct {
	aggregate aggregates.Aggregate
	args      []Expression
	state     any
}

var _ AggregateExpression = &aggregateSubExpression{}

func (e *aggregateSubExpression) Step(ctx ExpressionContext, row shared.Row) error {
	var values []any
	for _, arg := range e.args {
		value, err := arg.ValueFrom(ctx, row)
		if err != nil {
			return err
		}

		values = append(values, value)
	}

	newState, err := e.aggregate.Step(e.state, values)
	if err != nil {
		return err
	}

	e.state = newState
	return nil
}

func (e *aggregateSubExpression) Done(ctx ExpressionContext) (any, error) {
	return e.aggregate.Done(e.state)
}

type nonAggregateExpression struct {
	expression Expression
	state      any
}

var _ AggregateExpression = &nonAggregateExpression{}

func (e *nonAggregateExpression) Step(ctx ExpressionContext, row shared.Row) error {
	value, err := e.expression.ValueFrom(ctx, row)
	if err != nil {
		return err
	}

	e.state = value
	return nil
}

func (e *nonAggregateExpression) Done(ctx ExpressionContext) (any, error) {
	return e.state, nil
}
