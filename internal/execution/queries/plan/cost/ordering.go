package cost

import (
	"math"

	"github.com/efritz/gostgres/internal/shared/impls"
)

var (
	sortCostPerRow        = impls.ResourceCost{Memory: 0.1}
	sortCostPerComparison = impls.ResourceCost{CPU: 0.1}
)

func ApplySortToCost(innerCost impls.NodeCost) impls.NodeCost {
	// Sorting reads the entire inner relation on startup
	cost := MaterializeCost(innerCost)

	// Add the cost of storing each row in-memory and performing the comparison sort
	n := float64(innerCost.Statistics.RowCount)
	cost.FixedCost = cost.FixedCost.Add(sortCostPerRow.ScaleUniform(n))
	cost.FixedCost = cost.FixedCost.Add(sortCostPerComparison.ScaleUniform(n * math.Log2(n)))

	return cost
}
