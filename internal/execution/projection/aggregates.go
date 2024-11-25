package projection

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type aggregateExpressionFactory struct {
	projectedExpressions []ProjectedExpression
}

var _ impls.AggregateExpressionFactory = &aggregateExpressionFactory{}

func NewAggregateFactory(projectedExpressions []ProjectedExpression) impls.AggregateExpressionFactory {
	return &aggregateExpressionFactory{
		projectedExpressions: projectedExpressions,
	}
}

func (f *aggregateExpressionFactory) String() string {
	return fmt.Sprintf("{%s}", serializeProjectedExpressions(f.projectedExpressions))
}

func (f *aggregateExpressionFactory) Fields() []fields.Field {
	return fieldsFromProjectedExpressions("", f.projectedExpressions)
}

func (f *aggregateExpressionFactory) Create(ctx impls.ExecutionContext) ([]impls.AggregateExpression, error) {
	var aggregateExpressions []impls.AggregateExpression
	for _, projectedExpression := range f.projectedExpressions {
		aggregateExpressions = append(aggregateExpressions, expressions.AsAggregate(ctx, projectedExpression.Expression))
	}

	return aggregateExpressions, nil
}

func (f *aggregateExpressionFactory) Optimize(ctx impls.OptimizationContext) {
	for i := range f.projectedExpressions {
		f.projectedExpressions[i].Expression = f.projectedExpressions[i].Expression.Fold()
	}
}
