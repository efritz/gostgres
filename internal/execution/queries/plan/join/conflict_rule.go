package join

type ConflictRule struct {
	antecedent uint // if these relations are present,
	consequent uint // then these relations must also be present
}

func NewConflictRule(antecedent, consequent uint) ConflictRule {
	return ConflictRule{
		antecedent: antecedent,
		consequent: consequent,
	}
}

func (r ConflictRule) Applies(s1, s2 uint) bool {
	return !overlaps(union(s1, s2), r.antecedent) || isSubset(r.consequent, union(s1, s2))
}
