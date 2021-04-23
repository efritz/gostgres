package nodes

import (
	"fmt"
	"strings"

	"github.com/efritz/gostgres/internal/shared"
)

func hashVisitor(hash map[string][]sourcedRow, index int) VisitorFunc {
	return func(row shared.Row) (bool, error) {
		key := hashValues(row.Values)

		hash[key] = append(hash[key], sourcedRow{
			index: index,
			row:   row,
		})

		return true, nil
	}
}

func hashValues(values []interface{}) string {
	strValues := make([]string, 0, len(values))
	for _, value := range values {
		strValues = append(strValues, fmt.Sprintf("%v", value))
	}

	return strings.Join(strValues, ":")
}
