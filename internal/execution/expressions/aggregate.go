package expressions

import (
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
)

func AsAggregate(ctx impls.ExecutionContext, e impls.Expression) (impls.AggregateExpression, error) {
	var (
		results        []*constantExpression
		subExpressions []impls.AggregateExpression
	)

	outerExpression, err := e.Map(func(e impls.Expression) (impls.Expression, error) {
		f, ok := e.(*functionExpression)
		if !ok {
			return e, nil
		}

		aggregate, ok := ctx.Catalog.Aggregates.Get(f.name)
		if !ok {
			return e, nil
		}

		placeholder := &constantExpression{}
		results = append(results, placeholder)
		subExpressions = append(subExpressions, &aggregateSubExpression{aggregate: aggregate, args: f.args})
		return placeholder, nil
	})
	if err != nil {
		return nil, err
	}

	if len(subExpressions) > 0 {
		return &explodedAggregateExpression{
			results:         results,
			subExpressions:  subExpressions,
			outerExpression: outerExpression,
		}, nil
	}

	return &nonAggregateExpression{
		expression: outerExpression,
	}, nil
}

type explodedAggregateExpression struct {
	results         []*constantExpression
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

		e.results[i].value = value
	}

	return e.outerExpression.ValueFrom(ctx, rows.Row{})
}

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
