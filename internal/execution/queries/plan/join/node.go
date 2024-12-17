package join

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/execution/queries/nodes/join"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/execution/queries/plan/util"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type JoinNode interface {
	fmt.Stringer
	ConvertRightJoinsToLeftJoins() JoinNode
	BuildDescriptor(b DescriptorSetBuilder) uint
	plan.LogicalNode
}

//
//

type joinNodeLeaf struct {
	relation plan.LogicalNode
}

func NewJoinLeafNode(relation plan.LogicalNode) JoinNode {
	return &joinNodeLeaf{relation: relation}
}

func (n *joinNodeLeaf) String() string {
	return n.relation.Name()
}

func (n *joinNodeLeaf) ConvertRightJoinsToLeftJoins() JoinNode {
	return n
}

func (n *joinNodeLeaf) BuildDescriptor(b DescriptorSetBuilder) uint {
	return bit(b.AddRelation(n.relation))
}

//
//

type joinNodeInternal struct {
	left     JoinNode
	right    JoinNode
	operator JoinOperator
	strategy logicalJoinStrategy
}

type JoinOperator struct {
	JoinType  join.JoinType
	Condition impls.Expression
}

func NewJoinInternalNode(left, right JoinNode, operator JoinOperator) JoinNode {
	return &joinNodeInternal{
		left:     left,
		right:    right,
		operator: operator,
	}
}

func (n *joinNodeInternal) String() string {
	return fmt.Sprintf("(%s) %s (%s) ON %s", n.left.String(), n.operator.JoinType, n.right.String(), n.operator.Condition)
}

func (n *joinNodeInternal) ConvertRightJoinsToLeftJoins() JoinNode {
	l := n.left.ConvertRightJoinsToLeftJoins()
	r := n.right.ConvertRightJoinsToLeftJoins()

	if n.operator.JoinType == join.JoinTypeRightOuter {
		return NewJoinInternalNode(r, l, JoinOperator{JoinType: join.JoinTypeLeftOuter, Condition: n.operator.Condition})
	}

	return NewJoinInternalNode(l, r, n.operator)
}

func (n *joinNodeInternal) BuildDescriptor(b DescriptorSetBuilder) uint {
	return b.AddNode(JoinDescriptor{
		originalLeftRelations:  n.left.BuildDescriptor(b),
		originalRightRelations: n.right.BuildDescriptor(b),
		operator:               n.operator,
		referencedRelations:    b.MaskReferencedTables(n.operator.Condition),
		isNullRejecting:        false, // TODO
	})
}

//
//
//
//
//
//

func (n *joinNodeLeaf) Name() string {
	return n.relation.Name()
}

func (n *joinNodeLeaf) Fields() []fields.Field {
	return n.relation.Fields()
}

func (n *joinNodeLeaf) AddFilter(ctx impls.OptimizationContext, filterExpression impls.Expression) {
	n.relation.AddFilter(ctx, filterExpression)
}

func (n *joinNodeLeaf) AddOrder(ctx impls.OptimizationContext, orderExpression impls.OrderExpression) {
	n.relation.AddOrder(ctx, orderExpression)
}

func (n *joinNodeLeaf) Optimize(ctx impls.OptimizationContext) {
	n.relation.Optimize(ctx)
}

func (n *joinNodeLeaf) EstimateCost() plan.Cost {
	return n.relation.EstimateCost()
}

func (n *joinNodeLeaf) Filter() impls.Expression {
	return n.relation.Filter()
}

func (n *joinNodeLeaf) Ordering() impls.OrderExpression {
	return n.relation.Ordering()
}

func (n *joinNodeLeaf) SupportsMarkRestore() bool {
	return n.relation.SupportsMarkRestore()
}

func (n *joinNodeLeaf) Build() nodes.Node {
	return n.relation.Build()
}

//
//

func (n *joinNodeInternal) Name() string {
	return ""
}

func (n *joinNodeInternal) Fields() []fields.Field {
	return append(slices.Clone(n.left.Fields()), n.right.Fields()...)
}

func (n *joinNodeInternal) AddFilter(ctx impls.OptimizationContext, filterExpression impls.Expression) {
	// TODO - determine when this might inappropriately alter query semantics
	n.operator.Condition = expressions.UnionFilters(n.operator.Condition, filterExpression)
}

func (n *joinNodeInternal) AddOrder(ctx impls.OptimizationContext, orderExpression impls.OrderExpression) {
	util.LowerOrder(ctx, orderExpression, n.left, n.right)
}

func (n *joinNodeInternal) Optimize(ctx impls.OptimizationContext) {
	// NOTE: Outer fields depend on nested loop join strategy
	// Merge and hash joins won't have have LHS rows available to RHS

	if n.operator.Condition != nil {
		n.operator.Condition = n.operator.Condition.Fold()
		// TODO - determine when this might inappropriately alter query semantics
		util.LowerFilter(ctx, n.operator.Condition, n.left)
		util.LowerFilter(ctx.AddOuterFields(n.left.Fields()), n.operator.Condition, n.right)
	}

	// TODO:
	//   (1) Ensure we don't re-optimize for multiple levels of the same join tree;
	//   disable optimization until we hit leaves.
	//   (2) Select lowest cost alternative. Requires a system for estimating costs
	//   which we don't currently have.
	for _, v := range EnumerateValidJoinTrees(n) {
		_ = v
		// fmt.Printf("Considering alternative join: %s\n", v)
	}

	n.left.Optimize(ctx)
	n.right.Optimize(ctx.AddOuterFields(n.left.Fields()))

	// TODO - determine when this might inappropriately alter query semantics
	n.operator.Condition = expressions.FilterDifference(n.operator.Condition, expressions.UnionFilters(n.left.Filter(), n.right.Filter()))
	n.strategy = &logicalNestedLoopJoinStrategy{n: n}
}

func (n *joinNodeInternal) EstimateCost() plan.Cost {
	return plan.Cost{} // TODO
}

func (n *joinNodeInternal) Filter() impls.Expression {
	return expressions.UnionFilters(n.operator.Condition, n.left.Filter(), n.right.Filter())
}

func (n *joinNodeInternal) Ordering() impls.OrderExpression {
	if n.strategy == nil {
		panic("No strategy set - optimization required before ordering can be determined")
	}

	return n.strategy.Ordering()
}

func (n *joinNodeInternal) SupportsMarkRestore() bool {
	return false
}

func (n *joinNodeInternal) Build() nodes.Node {
	left := n.left.Build()
	right := n.right.Build()

	return nodes.NewJoin(
		left,
		right,
		n.operator.Condition,
		n.Fields(),
		n.strategy.Build(left, right, n.Fields()),
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
	n *joinNodeInternal
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
		s.n.operator.Condition,
		fields,
	)
}

//
//

type logicalMergeJoinStrategy struct {
	n     *joinNodeInternal
	pairs []join.EqualityPair
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
	n     *joinNodeInternal
	pairs []join.EqualityPair
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

func decomposeFilter(n *joinNodeInternal) (pairs []join.EqualityPair, _ bool) {
	if n.operator.Condition == nil {
		return nil, false
	}

	for _, expr := range expressions.Conjunctions(n.operator.Condition) {
		if comparisonType, left, right := expressions.IsComparison(expr); comparisonType == expressions.ComparisonTypeEquals {
			if bindsAllFields(n.left, left) && bindsAllFields(n.right, right) {
				pairs = append(pairs, join.EqualityPair{Left: left, Right: right})
				continue
			}

			if bindsAllFields(n.left, right) && bindsAllFields(n.right, left) {
				pairs = append(pairs, join.EqualityPair{Left: right, Right: left})
				continue
			}
		}

		return nil, false
	}

	return pairs, len(pairs) > 0
}

func bindsAllFields(n JoinNode, expr impls.Expression) bool {
	for _, field := range expressions.Fields(expr) {
		if _, err := fields.FindMatchingFieldIndex(field, n.Fields()); err != nil {
			return false
		}
	}

	return true
}
