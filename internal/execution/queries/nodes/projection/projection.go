package projection

import (
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
)

func Project(ctx impls.ExecutionContext, row rows.Row, projection *projection.Projection) (rows.Row, error) {
	aliases := projection.Aliases()
	fields := projection.Fields()

	values := make([]any, 0, len(aliases))
	for _, field := range aliases {
		value, err := queries.Evaluate(ctx, field.Expression, row)
		if err != nil {
			return rows.Row{}, err
		}

		values = append(values, value)
	}

	return rows.NewRow(fields, values)
}
