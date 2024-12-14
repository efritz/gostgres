package join

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/queries/nodes/join"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type JoinNode interface {
	fmt.Stringer
	ConvertRightJoinsToLeftJoins() JoinNode
	BuildDescriptor(b DescriptorSetBuilder) uint
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
