package opt

import (
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/execution/queries/nodes/join"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalJoinNode struct {
	left     LogicalNode
	right    LogicalNode
	filter   impls.Expression
	fields   []fields.Field
	strategy logicalJoinStrategy
}

func NewJoin(left LogicalNode, right LogicalNode, condition impls.Expression) LogicalNode {
	return &logicalJoinNode{
		left:     left,
		right:    right,
		filter:   condition,
		fields:   append(left.Fields(), right.Fields()...),
		strategy: nil,
	}
}

func (n *logicalJoinNode) Name() string {
	return ""
}

func (n *logicalJoinNode) Fields() []fields.Field {
	return slices.Clone(n.fields)
}

func (n *logicalJoinNode) AddFilter(ctx impls.OptimizationContext, filterExpression impls.Expression) {
	n.filter = expressions.UnionFilters(n.filter, filterExpression)
}

func (n *logicalJoinNode) AddOrder(ctx impls.OptimizationContext, orderExpression impls.OrderExpression) {
	lowerOrder(ctx, orderExpression, n.left, n.right)
}

func (n *logicalJoinNode) Optimize(ctx impls.OptimizationContext) {
	// NOTE: Outer fields depend on nested loop join strategy
	// Merge and hash joins won't have have LHS rows available to RHS

	if n.filter != nil {
		n.filter = n.filter.Fold()
		lowerFilter(ctx, n.filter, n.left)
		lowerFilter(ctx.AddOuterFields(n.left.Fields()), n.filter, n.right)
	}

	n.left.Optimize(ctx)
	n.right.Optimize(ctx.AddOuterFields(n.left.Fields()))
	n.filter = expressions.FilterDifference(n.filter, expressions.UnionFilters(n.left.Filter(), n.right.Filter()))
	n.strategy = &logicalNestedLoopJoinStrategy{n: n}
}

func (n *logicalJoinNode) Filter() impls.Expression {
	return expressions.UnionFilters(n.filter, n.left.Filter(), n.right.Filter())
}

func (n *logicalJoinNode) Ordering() impls.OrderExpression {
	if n.strategy == nil {
		panic("No strategy set - optimization required before ordering can be determined")
	}

	return n.strategy.Ordering()
}

func (n *logicalJoinNode) SupportsMarkRestore() bool {
	return false
}

func (n *logicalJoinNode) Build() nodes.Node {
	left := n.left.Build()
	right := n.right.Build()

	return nodes.NewJoin(
		left,
		right,
		n.filter,
		n.fields,
		n.strategy.Build(left, right, n.fields),
	)
}

//
//

type logicalJoinStrategy interface {
	Ordering() impls.OrderExpression
	Build(left nodes.Node, right nodes.Node, fields []fields.Field) nodes.JoinStrategy
}

//
//

type logicalNestedLoopJoinStrategy struct {
	n *logicalJoinNode
}

var _ logicalJoinStrategy = &logicalNestedLoopJoinStrategy{}

func (s *logicalNestedLoopJoinStrategy) Ordering() impls.OrderExpression {
	leftOrdering := s.n.left.Ordering()
	if leftOrdering == nil {
		return nil
	}

	rightOrdering := s.n.right.Ordering()
	if rightOrdering == nil {
		return leftOrdering
	}

	return expressions.NewOrderExpression(append(leftOrdering.Expressions(), rightOrdering.Expressions()...))
}

func (s *logicalNestedLoopJoinStrategy) Build(left nodes.Node, right nodes.Node, fields []fields.Field) nodes.JoinStrategy {
	return join.NewNestedLoopJoinStrategy(
		left,
		right,
		s.n.filter,
		fields,
	)
}

//
//

type logicalMergeJoinStrategy struct {
	n     *logicalJoinNode
	pairs []nodes.EqualityPair
}

var _ logicalJoinStrategy = &logicalMergeJoinStrategy{}

func (s *logicalMergeJoinStrategy) Ordering() impls.OrderExpression {
	// TODO - can add right fields as well?
	return s.n.left.Ordering()
}

func (s *logicalMergeJoinStrategy) Build(left nodes.Node, right nodes.Node, fields []fields.Field) nodes.JoinStrategy {
	// return &mergeJoinStrategy{
	// 	n:      n,
	// 	pairs:  s.pairs,
	// 	fields: n.fields,
	// }
	return join.NewMergeJoinStrategy(
		left,
		right,
		s.pairs,
		fields,
	)
}

//
//

type logicalHashJoinStrategy struct {
	n     *logicalJoinNode
	pairs []nodes.EqualityPair
}

var _ logicalJoinStrategy = &logicalHashJoinStrategy{}

func (s *logicalHashJoinStrategy) Ordering() impls.OrderExpression {
	return s.n.left.Ordering()
}

func (s *logicalHashJoinStrategy) Build(left nodes.Node, right nodes.Node, fields []fields.Field) nodes.JoinStrategy {
	return join.NewHashJoinStrategy(
		left,
		right,
		s.pairs,
		fields,
	)
}

//
//

// const (
// 	EnableHashJoins  = false
// 	EnableMergeJoins = false
// )

// func selectJoinStrategy(ctx impls.OptimizationContext, n *joinNode) joinStrategy {
// 	if pairs, ok := decomposeFilter(n); ok {
// 		if EnableMergeJoins {
// 			// if orderable?
// 			// if n.right.SupportsMarkRestore()

// 			// TODO - HACK!
// 			var lefts, rights []impls.ExpressionWithDirection
// 			for _, p := range pairs {
// 				lefts = append(lefts, impls.ExpressionWithDirection{Expression: p.left})
// 				rights = append(rights, impls.ExpressionWithDirection{Expression: p.right})
// 			}
// 			n.left = order.NewOrder(n.left, expressions.NewOrderExpression(lefts))
// 			n.left.Optimize(ctx)
// 			n.right = order.NewOrder(n.right, expressions.NewOrderExpression(rights))
// 			n.right.Optimize(ctx)

// 			return &mergeJoinStrategy{
// 				n:     n,
// 				pairs: pairs,
// 			}
// 		}

// 		if EnableHashJoins {
// 			return &hashJoinStrategy{
// 				n:     n,
// 				pairs: pairs,
// 			}
// 		}
// 	}

// 	// if n.filter != nil {
// 	// filter.LowerFilter(n.filter, n.left.Fields(), n.right)
// 	// n.right.AddFilter(n.filter)
// 	// n.right.Optimize(ctx)
// 	// }

// 	n.left.Optimize(ctx)
// 	n.right.Optimize(ctx)

// 	return &nestedLoopJoinStrategy{n: n}
// }

func decomposeFilter(n *logicalJoinNode) (pairs []nodes.EqualityPair, _ bool) {
	if n.filter == nil {
		return nil, false
	}

	for _, expr := range expressions.Conjunctions(n.filter) {
		if comparisonType, left, right := expressions.IsComparison(expr); comparisonType == expressions.ComparisonTypeEquals {
			if bindsAllFields(n.left, left) && bindsAllFields(n.right, right) {
				pairs = append(pairs, nodes.EqualityPair{Left: left, Right: right})
				continue
			}

			if bindsAllFields(n.left, right) && bindsAllFields(n.right, left) {
				pairs = append(pairs, nodes.EqualityPair{Left: right, Right: left})
				continue
			}
		}

		return nil, false
	}

	return pairs, len(pairs) > 0
}

func bindsAllFields(n LogicalNode, expr impls.Expression) bool {
	for _, field := range expressions.Fields(expr) {
		if _, err := fields.FindMatchingFieldIndex(field, n.Fields()); err != nil {
			return false
		}
	}

	return true
}
