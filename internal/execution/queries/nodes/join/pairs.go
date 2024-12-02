package join

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
)

type EqualityPair struct {
	Left  impls.Expression
	Right impls.Expression
}

var leftOfPair = func(pair EqualityPair) impls.Expression { return pair.Left }
var rightOfPair = func(pair EqualityPair) impls.Expression { return pair.Right }

func evaluatePair(ctx impls.ExecutionContext, pairs []EqualityPair, expression func(EqualityPair) impls.Expression, row rows.Row) (values []any, _ error) {
	for _, pair := range pairs {
		value, err := queries.Evaluate(ctx, expression(pair), row)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}
