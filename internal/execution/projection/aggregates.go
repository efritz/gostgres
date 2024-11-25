package projection

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type AggregateProjectionFactory struct {
	projectedExpressions []ProjectedExpression
}

func NewAggregateFactory(projectedExpressions []ProjectedExpression) *AggregateProjectionFactory {
	return &AggregateProjectionFactory{
		projectedExpressions: projectedExpressions,
	}
}

func (f *AggregateProjectionFactory) String() string {
	return fmt.Sprintf("{%s}", serializeProjectedExpressions(f.projectedExpressions))
}

func (f *AggregateProjectionFactory) Fields() []fields.Field {
	return fieldsFromProjectedExpressions("", f.projectedExpressions)
}

func (f *AggregateProjectionFactory) Create(ctx impls.ExecutionContext) ([]impls.AggregateExpression, error) {
	var aggregateExpressions []impls.AggregateExpression
	for _, projectedExpression := range f.projectedExpressions {
		aggregateExpressions = append(aggregateExpressions, expressions.AsAggregate(ctx, projectedExpression.Expression))
	}

	return aggregateExpressions, nil
}

func (f *AggregateProjectionFactory) Optimize(ctx impls.OptimizationContext) {
	for i := range f.projectedExpressions {
		f.projectedExpressions[i].Expression = f.projectedExpressions[i].Expression.Fold()
	}
}
