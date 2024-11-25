package expressions

import (
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
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
