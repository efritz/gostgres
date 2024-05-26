package rows

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/fields"
)

func refineTypes(refinableFields []fields.Field, values []any) ([]fields.Field, []any, error) {
	if len(refinableFields) != len(values) {
		return nil, nil, fmt.Errorf("unexpected number of columns")
	}

	refinedFields := make([]fields.Field, 0, len(refinableFields))
	refinedValues := make([]any, 0, len(values))

	for i, field := range refinableFields {
		refinedType, refinedValue, ok := field.Type().Refine(values[i])
		if !ok {
			return nil, nil, fmt.Errorf("type error (%v is not %s)", values[i], field.Type())
		}

		refinedFields = append(refinedFields, field.WithType(refinedType))
		refinedValues = append(refinedValues, refinedValue)
	}

	return refinedFields, refinedValues, nil
}
