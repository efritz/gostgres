package join

import (
	"math/bits"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type DescriptorSet interface {
	NumRelations() int
	Relation(mask uint) plan.LogicalNode
	ForEachDescriptor(f func(JoinDescriptor))
}

type DescriptorSetBuilder interface {
	AddRelation(relation plan.LogicalNode) uint
	AddNode(node JoinDescriptor) uint
	MaskReferencedTables(condition impls.Expression) uint
}

type descriptorSet struct {
	relations   []plan.LogicalNode
	descriptors []JoinDescriptor
}

func NewDescriptorSet(node JoinNode) DescriptorSet {
	ds := &descriptorSet{}
	_ = node.BuildDescriptor(ds)
	return ds
}

func (ds *descriptorSet) NumRelations() int {
	return len(ds.relations)
}

func (ds *descriptorSet) Relation(mask uint) plan.LogicalNode {
	return ds.relations[bits.TrailingZeros(mask)]
}

func (ds *descriptorSet) ForEachDescriptor(f func(JoinDescriptor)) {
	for _, d := range ds.descriptors {
		f(d)
	}
}

func (ds *descriptorSet) AddRelation(relation plan.LogicalNode) uint {
	i := len(ds.relations)
	ds.relations = append(ds.relations, relation)
	return uint(i)
}

func (ds *descriptorSet) AddNode(node JoinDescriptor) uint {
	ds.descriptors = append(ds.descriptors, node)
	return union(node.originalLeftRelations, node.originalRightRelations)
}

func (ds *descriptorSet) MaskReferencedTables(condition impls.Expression) uint {
	conditionFields := expressions.Fields(condition)

	mask := uint(0)
	for i, r := range ds.relations {
		for _, f := range conditionFields {
			if _, err := fields.FindMatchingFieldIndex(f, r.Fields()); err == nil {
				mask = set(mask, uint(i))
			}
		}
	}

	return mask
}
