package cost

import "github.com/efritz/gostgres/internal/shared/impls"

var projectionEvaluationCostPerRow = impls.ResourceCost{CPU: 0.01}

func ApplyProjectionToCost(innerCost impls.NodeCost) impls.NodeCost {
	cost := innerCost

	// Evaluate a projection expression for every row
	n := float64(innerCost.Statistics.RowCount)
	cost.VariableCost = cost.VariableCost.Add(projectionEvaluationCostPerRow.ScaleUniform(n))

	// TODO - update statistics
	return cost
}
