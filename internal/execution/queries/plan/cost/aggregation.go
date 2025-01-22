package cost

import "github.com/efritz/gostgres/internal/shared/impls"

var (
	buildAggregateTableCostPerRow    = impls.ResourceCost{CPU: 0.1}
	buildAggregateTableCostPerBucket = impls.ResourceCost{Memory: 0.1}
)

func ApplyAggregationToCost(innerCost impls.NodeCost) impls.NodeCost {
	// TODO - will need additional information, group expressions, etc
	estimatedNumBuckets := EstimateDistinctCount(innerCost.Statistics.RowCount)

	// Aggregation reads the entire inner relation on startup
	cost := MaterializeCost(innerCost)

	// Add the cost of creating aggregation buckets and hashing all rows from the inner relation
	n := float64(innerCost.Statistics.RowCount)
	cost.FixedCost = cost.FixedCost.Add(buildAggregateTableCostPerRow.ScaleUniform(n))
	cost.FixedCost = cost.FixedCost.Add(buildAggregateTableCostPerBucket.ScaleUniform(float64(estimatedNumBuckets)))

	// One row is projected for each bucket
	cost.Statistics.RowCount = estimatedNumBuckets
	cost = ApplyProjectionToCost(cost)

	// TODO - update statistics
	return cost
}

func EstimateDistinctCount(estimatedRows int) int {
	return estimatedRows // TODO
}
