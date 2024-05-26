package expressions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchesOrderType(t *testing.T) {
	type pair struct {
		lVal any
		rVal any
	}
	for _, testCase := range []struct {
		comparisonType ComparisonType
		matching       []pair
		nonMatching    []pair
	}{
		{
			comparisonType: ComparisonTypeEquals,
			matching:       []pair{{int32(200), int32(200)}},
			nonMatching:    []pair{{int32(100), int32(200)}},
		},
		{
			comparisonType: ComparisonTypeDistinctFrom,
			matching:       []pair{{int32(100), int32(200)}, {nil, int32(200)}, {int32(200), nil}},
			nonMatching:    []pair{{int32(200), int32(200)}, {nil, nil}},
		},
		{
			comparisonType: ComparisonTypeLessThan,
			matching:       []pair{{int32(100), int32(200)}},
			nonMatching:    []pair{{int32(200), int32(200)}, {int32(300), int32(200)}},
		},
		{
			comparisonType: ComparisonTypeLessThanEquals,
			matching:       []pair{{int32(100), int32(200)}, {int32(200), int32(200)}},
			nonMatching:    []pair{{int32(300), int32(200)}},
		},
		{
			comparisonType: ComparisonTypeGreaterThan,
			matching:       []pair{{int32(200), int32(100)}},
			nonMatching:    []pair{{int32(200), int32(200)}, {int32(100), int32(200)}},
		},
		{
			comparisonType: ComparisonTypeGreaterThanEquals,
			matching:       []pair{{int32(200), int32(100)}, {int32(200), int32(200)}},
			nonMatching:    []pair{{int32(100), int32(200)}},
		},
	} {
		for _, pair := range testCase.matching {
			name := fmt.Sprintf(
				"%v %s %v",
				pair.lVal,
				testCase.comparisonType.String(),
				pair.rVal,
			)

			t.Run(name, func(t *testing.T) {
				matches, err := testCase.comparisonType.MatchesOrderType(pair.lVal, pair.rVal)
				require.NoError(t, err)
				assert.True(t, matches)
			})
		}

		for _, pair := range testCase.nonMatching {
			name := fmt.Sprintf(
				"%v %s %v",
				pair.lVal,
				testCase.comparisonType.String(),
				pair.rVal,
			)

			t.Run(name, func(t *testing.T) {
				matches, err := testCase.comparisonType.MatchesOrderType(pair.lVal, pair.rVal)
				require.NoError(t, err)
				assert.False(t, matches)
			})
		}
	}
}
