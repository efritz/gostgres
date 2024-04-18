package nodes

import (
	"io"

	"github.com/efritz/gostgres/internal/expressions"
)

type accessStrategy interface {
	Serialize(w io.Writer, indentationLevel int)
	Filter() expressions.Expression
	Ordering() OrderExpression
	Scanner(ctx ScanContext) (Scanner, error)
}

func selectAccessStrategy(
	table *Table,
	filter expressions.Expression,
	order OrderExpression,
) accessStrategy {
	var candidates []accessStrategy
	for _, index := range table.indexes {
		if ias, ok := canSelectIndex(table, index, filter, order); ok {
			candidates = append(candidates, ias)
		}
	}

	maxScore := 0
	bestStrategy := NewTableAccessStrategy(table)

	for _, index := range candidates {
		indexCost := 0
		if subsumesOrder(order, index.Ordering()) {
			indexCost += 100
		}

		if filter != nil {
			oldLen := len(filter.Conjunctions())
			newLen := 0
			if remainingFilter := filterDifference(filter, index.Filter()); remainingFilter != nil {
				newLen = len(remainingFilter.Conjunctions())
			}

			indexCost += (oldLen - newLen) * 10
		}

		if indexCost > maxScore {
			bestStrategy = index
			maxScore = indexCost
		}
	}

	return bestStrategy
}

func canSelectIndex(
	table *Table,
	index BaseIndex,
	filter expressions.Expression,
	order OrderExpression,
) (accessStrategy, bool) {
	if indexFilter := index.Filter(); indexFilter != nil {
		if filter == nil {
			return nil, false
		}

		// TODO - need to do a more tight "subsumes" check
		for _, v := range indexFilter.Conjunctions() {
			if diff := filterDifference(v, filter); diff != nil && len(diff.Conjunctions()) >= len(filter.Conjunctions()) {
				return nil, false
			}
		}
	}

	if ias, ok := canSelectHashIndex(table, index, filter); ok {
		return ias, true
	}

	if ias, ok := canSelectBtreeIndex(table, index, filter, order); ok {
		return ias, true
	}

	return nil, false
}

type HashExpressioner interface {
	HashExpression() expressions.Expression
}

func (i *hashIndex) HashExpression() expressions.Expression {
	return i.expression
}

func canSelectHashIndex(
	table *Table,
	index BaseIndex,
	filter expressions.Expression,
) (accessStrategy, bool) {
	if filter == nil {
		return nil, false
	}

	hashIndex, ok := index.(Index[hashIndexScanOptions])
	if !ok {
		return nil, false
	}
	hashExpressioner, ok := hashIndex.Unwrap().(HashExpressioner)
	if !ok {
		return nil, false
	}
	hashExpression := hashExpressioner.HashExpression()

	for _, conjunction := range filter.Conjunctions() {
		if comparisonType, left, right := expressions.IsComparison(conjunction); comparisonType == expressions.ComparisonTypeEquals {
			if left.Equal(hashExpression) {
				return NewIndexAccessStrategy(table, hashIndex, hashIndexScanOptions{expression: right}), true
			}

			if right.Equal(hashExpression) {
				return NewIndexAccessStrategy(table, hashIndex, hashIndexScanOptions{expression: left}), true
			}
		}
	}

	return nil, false
}

type BTreeExpressioner interface {
	Expressions() []ExpressionWithDirection
}

func (i *btreeIndex) Expressions() []ExpressionWithDirection {
	return i.expressions
}

func canSelectBtreeIndex(
	table *Table,
	index BaseIndex,
	filter expressions.Expression,
	order OrderExpression,
) (accessStrategy, bool) {
	btreeIndex, ok := index.(Index[btreeIndexScanOptions])
	if !ok {
		return nil, false
	}
	btreeExpressioner, ok := btreeIndex.Unwrap().(BTreeExpressioner)
	if !ok {
		return nil, false
	}
	btreeExpressions := btreeExpressioner.Expressions()

	scanDirection := scanDirection(order, btreeExpressions)
	if scanDirection == ScanDirectionUnknown {
		scanDirection = ScanDirectionForward
	}

	lowerBounds, upperBounds := extractBounds(filter, btreeExpressions)

	return NewIndexAccessStrategy(table, btreeIndex, btreeIndexScanOptions{
		scanDirection: scanDirection,
		lowerBounds:   lowerBounds,
		upperBounds:   upperBounds,
	}), true
}

func scanDirection(order OrderExpression, indexDirections []ExpressionWithDirection) ScanDirection {
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

func extractBounds(filter expressions.Expression, indexedExprs []ExpressionWithDirection) (lowerBounds, upperBounds [][]scanBound) {
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
