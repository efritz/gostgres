package indexes

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type TableIndexer interface {
	Row(tid int) (shared.Row, bool)
}

func CanSelectHashIndex(
	table TableIndexer,
	index BaseIndex,
	filter expressions.Expression,
) (_ Index[hashIndexScanOptions], opts hashIndexScanOptions, _ bool) {
	// TODO - deduplicate
	if indexFilter := index.Filter(); indexFilter != nil {
		if filter == nil {
			return nil, opts, false
		}

		// TODO - need to do a more tight "subsumes" check
		for _, v := range indexFilter.Conjunctions() {
			if diff := filterDifference(v, filter); diff != nil && len(diff.Conjunctions()) >= len(filter.Conjunctions()) {
				return nil, opts, false
			}
		}
	}

	//
	//

	if filter == nil {
		return nil, opts, false
	}

	hashIndex, ok := index.(Index[hashIndexScanOptions])
	if !ok {
		return nil, opts, false
	}
	hashExpressioner, ok := hashIndex.Unwrap().(HashExpressioner)
	if !ok {
		return nil, opts, false
	}
	hashExpression := hashExpressioner.HashExpression()

	for _, conjunction := range filter.Conjunctions() {
		if comparisonType, left, right := expressions.IsComparison(conjunction); comparisonType == expressions.ComparisonTypeEquals {
			if left.Equal(hashExpression) {
				return hashIndex, hashIndexScanOptions{expression: right}, true
			}

			if right.Equal(hashExpression) {
				return hashIndex, hashIndexScanOptions{expression: left}, true
			}
		}
	}

	return nil, opts, false
}

func CanSelectBtreeIndex(
	table TableIndexer,
	index BaseIndex,
	filter expressions.Expression,
	order expressions.OrderExpression,
) (_ Index[btreeIndexScanOptions], opts btreeIndexScanOptions, _ bool) {
	// TODO - deduplicate
	if indexFilter := index.Filter(); indexFilter != nil {
		if filter == nil {
			return nil, opts, false
		}

		// TODO - need to do a more tight "subsumes" check
		for _, v := range indexFilter.Conjunctions() {
			if diff := filterDifference(v, filter); diff != nil && len(diff.Conjunctions()) >= len(filter.Conjunctions()) {
				return nil, opts, false
			}
		}
	}

	btreeIndex, ok := index.(Index[btreeIndexScanOptions])
	if !ok {
		return nil, opts, false
	}
	btreeExpressioner, ok := btreeIndex.Unwrap().(BTreeExpressioner)
	if !ok {
		return nil, opts, false
	}
	btreeExpressions := btreeExpressioner.Expressions()

	scanDirection := scanDirection(order, btreeExpressions)
	if scanDirection == ScanDirectionUnknown {
		scanDirection = ScanDirectionForward
	}

	lowerBounds, upperBounds := extractBounds(filter, btreeExpressions)

	opts = btreeIndexScanOptions{
		scanDirection: scanDirection,
		lowerBounds:   lowerBounds,
		upperBounds:   upperBounds,
	}
	return btreeIndex, opts, true
}

func scanDirection(order expressions.OrderExpression, indexDirections []expressions.ExpressionWithDirection) ScanDirection {
	if order == nil {
		return ScanDirectionUnknown
	}
	orderExpressions := order.Expressions()

	if len(orderExpressions) > len(indexDirections) {
		return ScanDirectionUnknown
	}

	var forward bool
	for i, orderExpr := range orderExpressions {
		if !orderExpr.Expression.Equal(indexDirections[i].Expression) {
			return ScanDirectionUnknown
		}

		matchesDirection := orderExpr.Reverse == indexDirections[i].Reverse

		if i == 0 {
			forward = matchesDirection
		} else if forward != matchesDirection {
			return ScanDirectionUnknown
		}
	}

	if !forward {
		return ScanDirectionBackward
	}

	return ScanDirectionForward
}

func extractBounds(filter expressions.Expression, indexedExprs []expressions.ExpressionWithDirection) (lowerBounds, upperBounds [][]scanBound) {
	if filter == nil {
		return nil, nil
	}

	conjunctions := filter.Conjunctions()

	for _, target := range indexedExprs {
		var (
			exprLowerBounds []scanBound
			exprUpperBounds []scanBound
		)

		for _, conjunction := range conjunctions {
			if comparisonType, left, right := expressions.IsComparison(conjunction); comparisonType != expressions.ComparisonTypeUnknown {
				if right.Equal(target.Expression) {
					left, right = right, left
					comparisonType = comparisonType.Flip()
				}

				if left.Equal(target.Expression) {
					switch comparisonType {
					case expressions.ComparisonTypeEquals:
						exprLowerBounds = append(exprLowerBounds, scanBound{expression: right, inclusive: true})
						exprUpperBounds = append(exprUpperBounds, scanBound{expression: right, inclusive: true})
					case expressions.ComparisonTypeLessThan:
						exprUpperBounds = append(exprUpperBounds, scanBound{expression: right, inclusive: false})
					case expressions.ComparisonTypeLessThanEquals:
						exprUpperBounds = append(exprUpperBounds, scanBound{expression: right, inclusive: true})
					case expressions.ComparisonTypeGreaterThan:
						exprLowerBounds = append(exprLowerBounds, scanBound{expression: right, inclusive: false})
					case expressions.ComparisonTypeGreaterThanEquals:
						exprLowerBounds = append(exprLowerBounds, scanBound{expression: right, inclusive: true})
					}
				}
			}
		}

		lowerBounds = append(lowerBounds, exprLowerBounds)
		upperBounds = append(upperBounds, exprUpperBounds)
	}

	prunedLowerBounds := lowerBounds[:0]
	for _, bound := range lowerBounds {
		if len(bound) == 0 {
			break
		}

		prunedLowerBounds = append(prunedLowerBounds, bound)
	}

	prunedUpperBounds := upperBounds[:0]
	for _, bound := range upperBounds {
		if len(bound) == 0 {
			break
		}

		prunedUpperBounds = append(prunedUpperBounds, bound)
	}

	return prunedLowerBounds, prunedUpperBounds
}

// TODO - deduplicate

func filterDifference(filter, childFilter expressions.Expression) expressions.Expression {
	return combineFilters(filter, childFilter, func(conjunctions, childConjunctions []expressions.Expression) {
		for i, f1 := range conjunctions {
			for _, f2 := range childConjunctions {
				if f1.Equal(f2) {
					conjunctions[i] = nil
					break
				}
			}
		}
	})
}

func combineFilters(filter, childFilter expressions.Expression, filterConjunctions func(conjunctions, childConjunctions []expressions.Expression)) expressions.Expression {
	if filter == nil {
		return nil
	}
	if childFilter == nil {
		return filter
	}

	conjunctions := filter.Conjunctions()
	filterConjunctions(conjunctions, childFilter.Conjunctions())
	return unionFilters(conjunctions...)
}

func unionFilters(filters ...expressions.Expression) expressions.Expression {
	var conjunctions []expressions.Expression
	for _, expression := range filters {
		if expression == nil {
			continue
		}

		conjunctions = append(conjunctions, expression.Conjunctions()...)
	}
	if len(conjunctions) == 0 {
		return nil
	}

	for i, c1 := range conjunctions {
		for j, c2 := range conjunctions {
			if c1 == nil || c2 == nil || j <= i {
				continue
			}

			if c1.Equal(c2) {
				conjunctions[j] = nil
			}
		}
	}

	filter := conjunctions[0]
	for _, conjunction := range conjunctions[1:] {
		if conjunction == nil {
			continue
		}

		filter = expressions.NewAnd(filter, conjunction)
	}

	return filter
}
