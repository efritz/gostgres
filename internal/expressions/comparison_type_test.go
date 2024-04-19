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
			matching:       []pair{{200, 200}},
			nonMatching:    []pair{{100, 200}},
		},
		{
			comparisonType: ComparisonTypeDistinctFrom,
			matching:       []pair{{100, 200}, {nil, 200}, {200, nil}},
			nonMatching:    []pair{{200, 200}, {nil, nil}},
		},
		{
			comparisonType: ComparisonTypeLessThan,
			matching:       []pair{{100, 200}},
			nonMatching:    []pair{{200, 200}, {300, 200}},
		},
		{
			comparisonType: ComparisonTypeLessThanEquals,
			matching:       []pair{{100, 200}, {200, 200}},
			nonMatching:    []pair{{300, 200}},
		},
		{
			comparisonType: ComparisonTypeGreaterThan,
			matching:       []pair{{200, 100}},
			nonMatching:    []pair{{200, 200}, {100, 200}},
		},
		{
			comparisonType: ComparisonTypeGreaterThanEquals,
			matching:       []pair{{200, 100}, {200, 200}},
			nonMatching:    []pair{{100, 200}},
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
