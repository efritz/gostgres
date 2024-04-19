package nodes

import (
	"fmt"
	"strings"
)

func hashValues(values []any) string {
	strValues := make([]string, 0, len(values))
	for _, value := range values {
		strValues = append(strValues, fmt.Sprintf("%v", value))
	}

	return strings.Join(strValues, ":")
}
