package cost

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/ordering"
	"github.com/efritz/gostgres/internal/shared/rows"
)

const defaultSelectivity = 0.7 // TODO

func estimateExpressionSelectivity(cond impls.Expression, rowCount int, columnStatistics []impls.ColumnStatistics) (s float64) {
	// fmt.Printf("ESTIMATING %s -> \n", cond)
	// defer func() { fmt.Printf("ESTIMATING %s -> %.12f\n", cond, s) }()

	if cond == nil {
		return 1.0
	}

	if expressions.IsConjunction(cond) {
		selectivity := 1.0
		for _, conjunction := range expressions.Conjunctions(cond) {
			selectivity *= estimateExpressionSelectivity(conjunction, rowCount, columnStatistics)
		}

		return selectivity
	}

	if expressions.IsDisjunction(cond) {
		selectivity := 1.0
		for _, disjunction := range expressions.Disjunctions(cond) {
			selectivity *= 1 - estimateExpressionSelectivity(disjunction, rowCount, columnStatistics)
		}

		return 1 - selectivity
	}

	if inner, ok := expressions.IsNegation(cond); ok {
		return 1 - estimateExpressionSelectivity(inner, rowCount, columnStatistics)
	}

	comparisonType, left, right := expressions.IsComparison(cond)
	if comparisonType == expressions.ComparisonTypeUnknown {
		return defaultSelectivity
	}

	if _, ok := findStatisticsForColumn(columnStatistics, right); ok {
		left, right = right, left
		comparisonType = comparisonType.Flip()
	}

	stats, ok := findStatisticsForColumn(columnStatistics, left)
	if !ok {
		return defaultSelectivity
	}

	value, err := right.ValueFrom(impls.EmptyExecutionContext, rows.Row{})
	if err != nil {
		return defaultSelectivity
	}

	switch comparisonType {
	case expressions.ComparisonTypeEquals:
		return estimateEqualitySelectivity(stats, rowCount, value)

	case
		expressions.ComparisonTypeLessThan,
		expressions.ComparisonTypeLessThanEquals,
		expressions.ComparisonTypeGreaterThan,
		expressions.ComparisonTypeGreaterThanEquals:
		return estimateComparisonSelectivity(stats, rowCount, comparisonType, value)
	}

	return defaultSelectivity
}

func findStatisticsForColumn(columnStatistics []impls.ColumnStatistics, expr impls.Expression) (impls.ColumnStatistics, bool) {
	for _, columnStats := range columnStatistics {
		// TODO - better way to determine field
		if expr.Equal(expressions.NewNamed(columnStats.Field)) {
			return columnStats, true
		}
	}

	return impls.ColumnStatistics{}, false
}

func estimateEqualitySelectivity(columnStatistics impls.ColumnStatistics, rowCount int, value any) float64 {
	nonNullFraction := 1 - columnStatistics.NullFraction

	for _, mcv := range columnStatistics.MostCommonValues {
		if ordering.CompareValues(mcv.Value, value) == ordering.OrderTypeEqual {
			return mcv.Frequency * nonNullFraction
		}
	}

	remainingFrequency := 1.0
	for _, mcv := range columnStatistics.MostCommonValues {
		remainingFrequency -= mcv.Frequency
	}

	distinctCount := columnStatistics.DistinctFraction * float64(rowCount)
	remainingDistinctCount := distinctCount - float64(len(columnStatistics.MostCommonValues))
	if remainingDistinctCount == 0 {
		return 0
	}

	return (remainingFrequency / remainingDistinctCount) * nonNullFraction
}

func estimateComparisonSelectivity(columnStatistics impls.ColumnStatistics, rowCount int, comparisonType expressions.ComparisonType, value any) float64 {
	bucketIndex := len(columnStatistics.HistogramBounds)
	for i, bound := range columnStatistics.HistogramBounds {
		if cmp := ordering.CompareValues(value, bound); cmp == ordering.OrderTypeBefore || cmp == ordering.OrderTypeEqual {
			bucketIndex = i
			break
		}
	}

	cumulativeHistogramFrequency := float64(bucketIndex) / float64(len(columnStatistics.HistogramBounds))
	if comparisonType == expressions.ComparisonTypeGreaterThan || comparisonType == expressions.ComparisonTypeGreaterThanEquals {
		cumulativeHistogramFrequency = 1 - cumulativeHistogramFrequency
	}

	mcvAdjustment := 0.0
	for _, mcv := range columnStatistics.MostCommonValues {
		if matchOrderWithComparisonType(mcv.Value, value, comparisonType) { // TODO - is this order correct?
			mcvAdjustment += mcv.Frequency
		}
	}

	nonNullFraction := 1 - columnStatistics.NullFraction
	return (cumulativeHistogramFrequency + mcvAdjustment) * nonNullFraction
}

func matchOrderWithComparisonType(left, right any, comparisonType expressions.ComparisonType) bool {
	switch ordering.CompareValues(left, right) {
	case ordering.OrderTypeBefore:
		return comparisonType == expressions.ComparisonTypeLessThan || comparisonType == expressions.ComparisonTypeLessThanEquals

	case ordering.OrderTypeEqual:
		return comparisonType == expressions.ComparisonTypeLessThanEquals || comparisonType == expressions.ComparisonTypeGreaterThanEquals

	case ordering.OrderTypeAfter:
		return comparisonType == expressions.ComparisonTypeGreaterThan || comparisonType == expressions.ComparisonTypeGreaterThanEquals
	}

	return false
}
