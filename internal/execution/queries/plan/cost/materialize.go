package cost

import "github.com/efritz/gostgres/internal/shared/impls"

func MaterializeCost(innerCost impls.NodeCost) impls.NodeCost {
	cost := innerCost

	// We pay the variable cost of the inner relation at startup
	cost.FixedCost = cost.FixedCost.Add(innerCost.VariableCost)

	// Reset the varaible cost so we're not double counting it
	cost.VariableCost = impls.ResourceCost{}

	return cost
}
