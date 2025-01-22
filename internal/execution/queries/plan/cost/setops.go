package cost

import "github.com/efritz/gostgres/internal/shared/impls"

var (
	uniqueSetCostPerRow    = impls.ResourceCost{CPU: 0.1}
	uniqueSetCostPerBucket = impls.ResourceCost{Memory: 0.1}
	rowSetCostPerRow       = impls.ResourceCost{CPU: 0.1}
	rowSetCostPerBucket    = impls.ResourceCost{Memory: 0.1}
)

func EstimateUnionCost(leftCost impls.NodeCost, rightCost impls.NodeCost, distinct bool) impls.NodeCost {
	estimatedLeftRows := float64(leftCost.Statistics.RowCount)
	estimatedRightRows := float64(rightCost.Statistics.RowCount)
	estimatedCombinedRows := estimatedLeftRows + estimatedRightRows
	estimatedDistinctRows := float64(EstimateDistinctCount(int(estimatedCombinedRows)))

	// On startup, we initialize both scanners
	fixedCost := leftCost.FixedCost.Add(rightCost.FixedCost)

	// In both cases, we scan all rows from both relations
	variableCost := leftCost.VariableCost.Add(rightCost.VariableCost)

	if !distinct {
		return impls.NodeCost{
			FixedCost:    fixedCost,
			VariableCost: variableCost,
			Statistics: impls.RelationStatistics{
				RowCount:         int(estimatedCombinedRows),
				ColumnStatistics: nil, // TODO - implement
			},
		}
	}

	variableCost = impls.SumCosts(
		variableCost,
		uniqueSetCostPerRow.ScaleUniform(estimatedCombinedRows),    // Cost to hash row and check unique set
		uniqueSetCostPerBucket.ScaleUniform(estimatedDistinctRows), // Cost to maintain running unique set
	)

	// TODO - update statistics
	return impls.NodeCost{
		FixedCost:    fixedCost,
		VariableCost: variableCost,
		Statistics: impls.RelationStatistics{
			RowCount:         int(estimatedDistinctRows),
			ColumnStatistics: nil, // TODO - implement
		},
	}
}

func EstimateIntersectCost(leftCost impls.NodeCost, rightCost impls.NodeCost, distinct bool) impls.NodeCost {
	return estimateSetOpCost(leftCost, rightCost, distinct, intersectionCardinalityEstimator)
}

func EstimateExceptCost(leftCost impls.NodeCost, rightCost impls.NodeCost, distinct bool) impls.NodeCost {
	return estimateSetOpCost(leftCost, rightCost, distinct, exceptCardinalityEstimator)
}

func intersectionCardinalityEstimator(
	estimatedLeftRows int, estimatedRightRows int,
	estimatedLeftDistinctRows int, estimatedRightDistinctRows int,
) (total, distinct int) {
	estimatedIntersectingRows := float64(estimatedRightRows) // TODO - estimate
	estimatedDistinctIntersectingRows := float64(EstimateDistinctCount(int(estimatedIntersectingRows)))
	return int(estimatedIntersectingRows), int(estimatedDistinctIntersectingRows)
}

func exceptCardinalityEstimator(
	estimatedLeftRows int, estimatedRightRows int,
	estimatedLeftDistinctRows int, estimatedRightDistinctRows int,
) (total, distinct int) {
	estimatedIntersectingRows := float64(estimatedLeftRows) // TODO - estimate
	estimatedDistinctIntersectingRows := EstimateDistinctCount(int(estimatedIntersectingRows))
	return int(estimatedIntersectingRows), int(estimatedDistinctIntersectingRows)
}

//
//

type setOpCardinalityEstimatorFunc func(
	estimatedLeftRows int, estimatedRightRows int,
	estimatedLeftDistinctRows int, estimatedRightDistinctRows int,
) (total int, distinct int)

func estimateSetOpCost(
	leftCost impls.NodeCost,
	rightCost impls.NodeCost,
	distinct bool,
	estimateCardinality setOpCardinalityEstimatorFunc,
) impls.NodeCost {
	estimatedLeftRows := float64(leftCost.Statistics.RowCount)
	estimatedRightRows := float64(rightCost.Statistics.RowCount)
	estimatedLeftDistinctRows := float64(EstimateDistinctCount(int(estimatedLeftRows)))
	estimatedRightDistinctRows := float64(EstimateDistinctCount(int(estimatedRightRows)))

	estimatedResultRows, estimatedDistinctResultRows := estimateCardinality(
		int(estimatedLeftRows),
		int(estimatedRightRows),
		int(estimatedLeftDistinctRows),
		int(estimatedRightDistinctRows),
	)

	estimatedRows := float64(estimatedResultRows)
	if distinct {
		estimatedRows = float64(estimatedDistinctResultRows)
	}

	fixedCost := impls.SumCosts(
		leftCost.FixedCost.Add(rightCost.FixedCost),                  // Cost to initialize both scanners
		rightCost.VariableCost,                                       // Cost to read right relation
		rowSetCostPerRow.ScaleUniform(estimatedRightRows),            // Cost to hash row
		rowSetCostPerBucket.ScaleUniform(estimatedRightDistinctRows), // Cost to maintain initial row set entrie
	)

	variableCost := impls.SumCosts(
		leftCost.VariableCost,                                       // Cost to scan each row from left relation
		rowSetCostPerRow.ScaleUniform(estimatedLeftRows),            // Cost to hash row
		rowSetCostPerBucket.ScaleUniform(estimatedLeftDistinctRows), // Cost to maintain additional row set entries
	)

	// TODO - update statistics
	return impls.NodeCost{
		FixedCost:    fixedCost,
		VariableCost: variableCost,
		Statistics: impls.RelationStatistics{
			RowCount:         int(estimatedRows),
			ColumnStatistics: nil, // TODO - implement
		},
	}
}
