package joins

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/queries/order"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
)

type joinStrategy interface {
	Name() string
	Ordering() expressions.OrderExpression
	Scanner(ctx queries.Context) (scan.Scanner, error)
}

const (
	EnableHashJoins  = false
	EnableMergeJoins = false
)

func selectJoinStrategy(n *joinNode) joinStrategy {
	if pairs, ok := decomposeFilter(n); ok {
		if EnableMergeJoins {
			// if orderable?
			// if n.right.SupportsMarkRestore()

			// TODO - HACK!
			var lefts, rights []expressions.ExpressionWithDirection
			for _, p := range pairs {
				lefts = append(lefts, expressions.ExpressionWithDirection{Expression: p.left})
				rights = append(rights, expressions.ExpressionWithDirection{Expression: p.right})
			}
			n.left = order.NewOrder(n.left, expressions.NewOrderExpression(lefts))
			n.left.Optimize()
			n.right = order.NewOrder(n.right, expressions.NewOrderExpression(rights))
			n.right.Optimize()

			return &mergeJoinStrategy{
				n:     n,
				pairs: pairs,
			}
		}

		if EnableHashJoins {
			return &hashJoinStrategy{
				n:     n,
				pairs: pairs,
			}
		}
	}

	if n.filter != nil {
		n.right.AddFilter(n.filter)
		n.right.Optimize()
	}

	return &nestedLoopJoinStrategy{n: n}
}

func decomposeFilter(n *joinNode) (pairs []equalityPair, _ bool) {
	if n.filter == nil {
		return nil, false
	}

	for _, expr := range expressions.Conjunctions(n.filter) {
		if comparisonType, left, right := expressions.IsComparison(expr); comparisonType == expressions.ComparisonTypeEquals {
			if bindsAllFields(n.left, left) && bindsAllFields(n.right, right) {
				pairs = append(pairs, equalityPair{left: left, right: right})
				continue
			}

			if bindsAllFields(n.left, right) && bindsAllFields(n.right, left) {
				pairs = append(pairs, equalityPair{left: right, right: left})
				continue
			}
		}

		return nil, false
	}

	return pairs, len(pairs) > 0
}

func bindsAllFields(n queries.Node, expr expressions.Expression) bool {
	for _, field := range expressions.Fields(expr) {
		if _, err := shared.FindMatchingFieldIndex(field, n.Fields()); err != nil {
			return false
		}
	}

	return true
}

type equalityPair struct {
	left  expressions.Expression
	right expressions.Expression
}

var leftOfPair = func(pair equalityPair) expressions.Expression { return pair.left }
var rightOfPair = func(pair equalityPair) expressions.Expression { return pair.right }

func evaluatePair(ctx queries.Context, pairs []equalityPair, expression func(equalityPair) expressions.Expression, row shared.Row) (values []any, _ error) {
	for _, pair := range pairs {
		value, err := queries.Evaluate(ctx, expression(pair), row)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}
