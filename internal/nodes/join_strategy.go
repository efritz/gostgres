package nodes

import "github.com/efritz/gostgres/internal/expressions"

type joinStrategy interface {
	Name() string
	Ordering() OrderExpression
	Scanner(ctx ScanContext) (Scanner, error)
}

const (
	EnableHashJoins  = false
	EnableMergeJoins = false
)

func selectJoinStrategy(
	n *joinNode,
) joinStrategy {
	comparisonType, left, right := n.decomposeFilter()
	if comparisonType == expressions.ComparisonTypeEquals {
		if EnableMergeJoins {
			// if orderable?
			// if n.right.SupportsMarkRestore()
			n.left = NewOrder(n.left, NewOrderExpression([]ExpressionWithDirection{{Expression: left}})) // HACK!
			n.left.Optimize()
			n.right = NewOrder(n.right, NewOrderExpression([]ExpressionWithDirection{{Expression: right}})) // HACK!
			n.right.Optimize()

			return &mergeJoinStrategy{
				n:     n,
				left:  left,
				right: right,
			}
		}

		if EnableHashJoins {
			return &hashJoinStrategy{
				n:     n,
				left:  left,
				right: right,
			}
		}
	}

	if n.filter != nil {
		n.right.AddFilter(n.filter)
		n.right.Optimize()
	}

	return &nestedLoopJoinStrategy{n: n}
}
