package join

import (
	"math/bits"
)

func EnumerateValidJoinTrees(node JoinNode) []JoinNode {
	descriptorSet := NewDescriptorSet(node)
	conflictRuleSet := NewConflictRuleSet(descriptorSet)
	n := uint(descriptorSet.NumRelations())

	valid := map[uint][]ProposedJoin{}
	for _, mask := range generateSubsetMasks(n) {
		if bits.OnesCount(mask) == 1 {
			valid[mask] = []ProposedJoin{newProposedJoinLeaf(mask)}
			continue
		}

		var trees []ProposedJoin
		for _, s1 := range generateSubsetMasksMatchingPattern(mask) {
			if s2 := intersect(complement(n, s1), mask); s2 != 0 {
				descriptorSet.ForEachDescriptor(func(descriptor JoinDescriptor) {
					if descriptor.Applicable(s1, s2) && conflictRuleSet.Applicable(descriptor, s1, s2) {
						trees = append(trees, newProposedJoinInternal(descriptor, s1, s2))

						if descriptor.IsCommutative() {
							trees = append(trees, newProposedJoinInternal(descriptor, s2, s1))
						}
					}
				})
			}
		}

		valid[mask] = trees
	}

	return reconstruct(descriptorSet, valid, all(n))
}
