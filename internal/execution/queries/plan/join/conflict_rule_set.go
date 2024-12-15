package join

type ConflictRuleSet interface {
	Applicable(op JoinDescriptor, s1, s2 uint) bool
}

type conflictRuleSet struct {
	rules map[JoinDescriptor][]ConflictRule
}

func NewConflictRuleSet(ds DescriptorSet) ConflictRuleSet {
	s := &conflictRuleSet{rules: map[JoinDescriptor][]ConflictRule{}}
	s.generateRules(ds)
	return s
}

func (s *conflictRuleSet) Applicable(op JoinDescriptor, s1, s2 uint) bool {
	for _, rule := range s.rules[op] {
		if !rule.Applies(s1, s2) {
			return false
		}
	}

	return true
}

func (s *conflictRuleSet) generateRules(ds DescriptorSet) {
	ds.ForEachDescriptor(func(desc1 JoinDescriptor) {
		ds.ForEachDescriptor(func(desc2 JoinDescriptor) {
			if desc1 != desc2 {
				s.generateRule(desc1, desc2)
			}
		})
	})
}

func (s *conflictRuleSet) generateRule(desc1, desc2 JoinDescriptor) {
	combined := union(desc2.originalLeftRelations, desc2.originalRightRelations)
	inLeftSubtree := isSubset(combined, desc1.originalLeftRelations)
	inRightSubtree := isSubset(combined, desc1.originalRightRelations)

	if (inLeftSubtree && !desc2.IsAssociative(desc1)) || (inRightSubtree && !desc1.IsRightAsscom(desc2)) {
		rule := NewConflictRule(desc2.originalRightRelations, coalesce(intersect(desc2.originalLeftRelations, desc2.referencedRelations), desc2.originalLeftRelations))
		s.rules[desc1] = append(s.rules[desc1], rule)
	}

	if (inRightSubtree && !desc1.IsAssociative(desc2)) || (inLeftSubtree && !desc2.IsLeftAsscom(desc1)) {
		rule := NewConflictRule(desc2.originalLeftRelations, coalesce(intersect(desc2.originalRightRelations, desc2.referencedRelations), desc2.originalRightRelations))
		s.rules[desc1] = append(s.rules[desc1], rule)
	}
}
