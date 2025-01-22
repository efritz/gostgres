package cost

import "github.com/efritz/gostgres/internal/shared/impls"

var (
	joinMergeCostPerRow  = impls.ResourceCost{CPU: 0.2}
	joinFilterCostPerRow = impls.ResourceCost{CPU: 0.1}
)

func EstimateNestedLoopJoinCost(
	leftCost impls.NodeCost,
	rightCost impls.NodeCost,
	joinSelectivity float64,
	hasCondition bool,
) impls.NodeCost {
	estimatedLeftRows := float64(leftCost.Statistics.RowCount)
	estimatedRightRows := float64(rightCost.Statistics.RowCount)
	estimatedCandidateRows := estimatedLeftRows * estimatedRightRows
	estimatedResultRows := estimatedCandidateRows * joinSelectivity

	costPerCandidateRow := joinMergeCostPerRow
	if hasCondition {
		costPerCandidateRow = costPerCandidateRow.Add(joinFilterCostPerRow)
	}

	// On startup, we only initialize the left scanner
	fixedCost := leftCost.FixedCost

	variableCost := impls.SumCosts(
		leftCost.VariableCost.ScaleUniform(estimatedLeftRows),       // Cost to scan each row from left relation
		rightCost.FixedCost.ScaleUniform(estimatedLeftRows),         // Cost to re-initialized right scanner for every row from left relation
		rightCost.VariableCost.ScaleUniform(estimatedCandidateRows), // Cost to scan each row from right relation
		costPerCandidateRow.ScaleUniform(estimatedCandidateRows),    // Cost to merge row pairs and evaluate the join condition
	)

	// TODO - update statistics
	return impls.NodeCost{
		FixedCost:    fixedCost,
		VariableCost: variableCost,
		Statistics: impls.RelationStatistics{
			RowCount:         int(estimatedResultRows),
			ColumnStatistics: nil, // TODO - implement
		},
	}
}
