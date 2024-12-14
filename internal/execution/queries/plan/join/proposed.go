package join

type ProposedJoin interface {
	Reconstruct(ds DescriptorSet, vaild map[uint][]ProposedJoin) []JoinNode
}

func reconstruct(ds DescriptorSet, vaild map[uint][]ProposedJoin, target uint) []JoinNode {
	var forest []JoinNode
	for _, b := range vaild[target] {
		forest = append(forest, b.Reconstruct(ds, vaild)...)
	}

	return forest
}

//
//

type proposedJoinLeaf struct {
	relation uint
}

func newProposedJoinLeaf(relation uint) proposedJoinLeaf {
	return proposedJoinLeaf{relation: relation}
}

func (b proposedJoinLeaf) Reconstruct(ds DescriptorSet, vaild map[uint][]ProposedJoin) []JoinNode {
	return []JoinNode{NewJoinLeafNode(ds.Relation(b.relation))}
}

//
//

type proposedJoinInternal struct {
	descriptor JoinDescriptor
	s1         uint
	s2         uint
}

func newProposedJoinInternal(descriptor JoinDescriptor, s1, s2 uint) proposedJoinInternal {
	return proposedJoinInternal{
		descriptor: descriptor,
		s1:         s1,
		s2:         s2,
	}
}

func (b proposedJoinInternal) Reconstruct(ds DescriptorSet, vaild map[uint][]ProposedJoin) []JoinNode {
	var forest []JoinNode
	for _, l := range reconstruct(ds, vaild, b.s1) {
		for _, r := range reconstruct(ds, vaild, b.s2) {
			forest = append(forest, NewJoinInternalNode(l, r, b.descriptor.operator))
		}
	}

	return forest
}
