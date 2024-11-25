package aggregates

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
)

type aggregateExpressionFactory struct {
	exprs []impls.Expression
}

var _ impls.AggregateExpressionFactory = &aggregateExpressionFactory{}

func NewAggregateFactory(exprs []impls.Expression) impls.AggregateExpressionFactory {
	return &aggregateExpressionFactory{
		exprs: exprs,
	}
}

// TODO - fields

func (f aggregateExpressionFactory) Create(ctx impls.ExecutionContext) ([]impls.AggregateExpression, error) {
	var aggregateExpressions []impls.AggregateExpression
	for _, expression := range f.exprs {
		aggregateExpressions = append(aggregateExpressions, asAggregate(ctx, expression))
	}

	return aggregateExpressions, nil
}

//
//

func asAggregate(ctx impls.ExecutionContext, e impls.Expression) impls.AggregateExpression {
	var (
		results        []expressions.ConstantPlaceholder
		subExpressions []impls.AggregateExpression
	)

	outerExpression, _ := e.Map(func(e impls.Expression) (impls.Expression, error) {
		if aggregate, args, ok := expressions.UnwrapAggregate(ctx, e); ok {
			placeholder := expressions.NewMutableConstant()
			results = append(results, placeholder)
			subExpressions = append(subExpressions, &aggregateSubExpression{aggregate: aggregate, args: args})
			return placeholder, nil
		}

		return e, nil
	})

	if len(results) > 0 {
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

//
//

type explodedAggregateExpression struct {
	results         []expressions.ConstantPlaceholder
	subExpressions  []impls.AggregateExpression
	outerExpression impls.Expression
}

var _ impls.AggregateExpression = &explodedAggregateExpression{}

func (e *explodedAggregateExpression) Step(ctx impls.ExecutionContext, row rows.Row) error {
	for _, subexpression := range e.subExpressions {
		if err := subexpression.Step(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func (e *explodedAggregateExpression) Done(ctx impls.ExecutionContext) (any, error) {
	for i, subExpression := range e.subExpressions {
		value, err := subExpression.Done(ctx)
		if err != nil {
			return nil, err
		}

		e.results[i].SetValue(value)
	}

	return e.outerExpression.ValueFrom(ctx, rows.Row{})
}

//
//

type aggregateSubExpression struct {
	aggregate impls.Aggregate
	args      []impls.Expression
	state     any
}

var _ impls.AggregateExpression = &aggregateSubExpression{}

func (e *aggregateSubExpression) Step(ctx impls.ExecutionContext, row rows.Row) error {
	var values []any
	for _, arg := range e.args {
		value, err := arg.ValueFrom(ctx, row)
		if err != nil {
			return err
		}

		values = append(values, value)
	}

	newState, err := e.aggregate.Step(ctx, e.state, values)
	if err != nil {
		return err
	}

	e.state = newState
	return nil
}

func (e *aggregateSubExpression) Done(ctx impls.ExecutionContext) (any, error) {
	return e.aggregate.Done(ctx, e.state)
}

//
//

type nonAggregateExpression struct {
	expression impls.Expression
	state      any
}

var _ impls.AggregateExpression = &nonAggregateExpression{}

func (e *nonAggregateExpression) Step(ctx impls.ExecutionContext, row rows.Row) error {
	value, err := e.expression.ValueFrom(ctx, row)
	if err != nil {
		return err
	}

	e.state = value
	return nil
}

func (e *nonAggregateExpression) Done(ctx impls.ExecutionContext) (any, error) {
	return e.state, nil
}
