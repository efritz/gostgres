package cost

import "github.com/efritz/gostgres/internal/shared/impls"

func AlterCostByLimitOffset(innerCost impls.NodeCost, limit, offset *int) impls.NodeCost {
	if limit == nil && offset == nil {
		return innerCost
	}

	cost := innerCost

	// Scale the variable cost of the inner relation by reading only limit + offset rows
	o := coalesce(offset, 0)
	l := coalesce(limit, innerCost.Statistics.RowCount-o)
	cost.VariableCost = cost.VariableCost.ScaleUniform(float64(l+o) / float64(innerCost.Statistics.RowCount))

	// Adjust number of output rows; this may be less than limit + offset, so we may end up
	// "smearing" the variable cost of the inner relation over fewer rows of the outer relation.
	cost.Statistics.RowCount = l

	// TODO - update statistics
	return cost
}
