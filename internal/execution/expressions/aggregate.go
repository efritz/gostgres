package expressions

import (
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
)

func UnwrapAggregate(ctx impls.ExecutionContext, expr impls.Expression) (impls.Aggregate, []impls.Expression, bool) {
	if f, ok := expr.(*functionExpression); ok {
		if aggregate, ok := ctx.Catalog.Aggregates.Get(f.name); ok {
			return aggregate, f.args, true
		}
	}

	return nil, nil, false
}

//
//

func PartitionAggregatedFieldReferences(
	ctx impls.ExpressionResolutionContext,
	exprs []impls.Expression,
	groupings []impls.Expression,
) (
	aggregatedFields []fields.Field,
	nonAggregatedFields []fields.Field,
	containsAggregate bool,
	_ error,
) {
	partitioner := &partitioner{
		groupings: groupings,
	}

	if err := partitioner.partitionExpressions(ctx, exprs); err != nil {
		return nil, nil, false, err
	}

	return partitioner.aggregatedFields, partitioner.nonAggregatedFields, partitioner.containsAggregate, nil
}

type partitioner struct {
	groupings []impls.Expression

	aggregatedFields    []fields.Field
	nonAggregatedFields []fields.Field
	containsAggregate   bool
}

func (p *partitioner) partitionExpressions(ctx impls.ExpressionResolutionContext, exprs []impls.Expression) error {
	for _, expr := range exprs {
		if err := p.partitionExpression(ctx, expr, false); err != nil {
			return err
		}
	}

	return nil
}

func (p *partitioner) partitionExpression(ctx impls.ExpressionResolutionContext, expr impls.Expression, inAggregate bool) error {
	for _, grouping := range p.groupings {
		if grouping.Equal(expr) {
			return nil
		}
	}

	switch expr := expr.(type) {
	case *functionExpression:
		_, isAggregate, _ := lookupFunction(ctx, expr.name)
		inAggregate = inAggregate || isAggregate
		p.containsAggregate = p.containsAggregate || isAggregate

	case NamedExpression:
		if inAggregate {
			p.aggregatedFields = append(p.aggregatedFields, expr.Field())
		} else {
			p.nonAggregatedFields = append(p.nonAggregatedFields, expr.Field())
		}
	}

	for _, expr := range expr.Children() {
		if err := p.partitionExpression(ctx, expr, inAggregate); err != nil {
			return err
		}
	}

	return nil
}

//
//

func AsAggregate(ctx impls.ExecutionContext, e impls.Expression) impls.AggregateExpression {
	var (
		results        []ConstantPlaceholder
		subExpressions []impls.AggregateExpression
	)

	outerExpression, _ := e.Map(func(e impls.Expression) (impls.Expression, error) {
		if aggregate, args, ok := UnwrapAggregate(ctx, e); ok {
			placeholder := NewMutableConstant()
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
	results         []ConstantPlaceholder
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
