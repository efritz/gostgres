package cost

import "github.com/efritz/gostgres/internal/shared/impls"

var filterEvaluationCostPerRow = impls.ResourceCost{CPU: 0.01}

func ApplyFilterToCost(innerCost impls.NodeCost, filter impls.Expression) impls.NodeCost {
	cost := innerCost

	// Evaluate a filter expression for every row
	n := float64(innerCost.Statistics.RowCount)
	cost.VariableCost = cost.VariableCost.Add(filterEvaluationCostPerRow.ScaleUniform(n))

	// Only rows selected by the filter are emitted
	selectivity := EstimateFilterSelectivity(filter, innerCost.Statistics)
	cost.Statistics.RowCount = int(n * selectivity)

	// TODO - update statistics
	return cost
}

// TODO - reduce uses outside of ApplyFilterToCost
func EstimateFilterSelectivity(filter impls.Expression, statistics impls.RelationStatistics) float64 {
	// TODO - collapse
	return estimateExpressionSelectivity(filter, statistics.RowCount, statistics.ColumnStatistics)
}
