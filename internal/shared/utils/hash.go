package utils

import (
	"fmt"
	"hash/fnv"
	"strings"
)

func Hash(value any) uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(fmt.Sprintf("%v", value)))
	return h.Sum64()
}

func HashSlice(values []any) string {
	strValues := make([]string, 0, len(values))
	for _, value := range values {
		strValues = append(strValues, fmt.Sprintf("%v", value))
	}

	return strings.Join(strValues, ":")
}
