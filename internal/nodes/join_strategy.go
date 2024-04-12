package nodes

import "github.com/efritz/gostgres/internal/expressions"

type joinStrategy interface {
	Name() string
	Ordering() OrderExpression
	Scanner() (Scanner, error)
}

func selectJoinStrategy(
	n *joinNode,
) joinStrategy {
	comparisonType, left, right := n.decomposeFilter()
	if comparisonType == expressions.ComparisonTypeEquals {
		if false {
			// if orderable?
			// if n.right.SupportsMarkRestore()
			n.left = NewOrder(n.left, NewOrderExpression([]ExpressionWithDirection{{Expression: left}}))    // HACK!
			n.right = NewOrder(n.right, NewOrderExpression([]ExpressionWithDirection{{Expression: right}})) // HACK!

			return &mergeJoinStrategy{
				n:     n,
				left:  left,
				right: right,
			}
		}

		return &hashJoinStrategy{
			n:     n,
			left:  left,
			right: right,
		}
	}

	return &nestedLoopJoinStrategy{n: n}
}
