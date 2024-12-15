package join

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes/join"
)

type JoinDescriptor struct {
	originalLeftRelations  uint
	originalRightRelations uint
	operator               JoinOperator
	referencedRelations    uint
	isNullRejecting        bool // TODO
}

func (n JoinDescriptor) Applicable(s1, s2 uint) bool {
	switch n.operator.JoinType {
	case join.JoinTypeInner, join.JoinTypeLeftOuter, join.JoinTypeFullOuter:
		// For these joins, we need all tables referenced in the predicate that are actually in the subtrees being joined
		tes := intersect(n.referencedRelations, union(n.originalLeftRelations, n.originalRightRelations))
		lTes := intersect(tes, n.originalLeftRelations)
		rTes := intersect(tes, n.originalRightRelations)
		return isSubset(lTes, s1) && isSubset(rTes, s2)
	}

	return true
}

func (op JoinDescriptor) IsCommutative() bool {
	switch op.operator.JoinType {
	case join.JoinTypeInner, join.JoinTypeFullOuter, join.JoinTypeCross:
		return true
	default:
		return false
	}
}

func (op1 JoinDescriptor) IsAssociative(op2 JoinDescriptor) bool {
	if op1.operator.JoinType == join.JoinTypeInner || op1.operator.JoinType == join.JoinTypeCross {
		return op2.operator.JoinType != join.JoinTypeFullOuter
	}

	if op1.operator.JoinType == join.JoinTypeLeftOuter {
		if op2.operator.JoinType == join.JoinTypeFullOuter {
			// TODO - footnote 1
			return op1.isNullRejecting
		}

		return true
	}

	if op1.operator.JoinType == join.JoinTypeFullOuter {
		if op2.operator.JoinType == join.JoinTypeLeftOuter {
			// TODO - footnote 2
			return op2.isNullRejecting
		}

		if op2.operator.JoinType == join.JoinTypeFullOuter {
			// TODO - footnote 3
			return op1.isNullRejecting && op2.isNullRejecting
		}
	}

	return false
}

func (op1 JoinDescriptor) IsLeftAsscom(op2 JoinDescriptor) bool {
	if op1.operator.JoinType == join.JoinTypeInner || op1.operator.JoinType == join.JoinTypeCross {
		return op2.operator.JoinType != join.JoinTypeFullOuter
	}

	if op1.operator.JoinType == join.JoinTypeLeftOuter {
		if op2.operator.JoinType == join.JoinTypeFullOuter {
			// TODO - footnote 1
			return op1.isNullRejecting
		}

		return true
	}

	if op1.operator.JoinType == join.JoinTypeFullOuter {
		if op2.operator.JoinType == join.JoinTypeLeftOuter {
			// TODO - footnote 2
			return op2.isNullRejecting
		}

		if op2.operator.JoinType == join.JoinTypeFullOuter {
			// TODO - footnote 3
			return op1.isNullRejecting && op2.isNullRejecting
		}
	}

	return false
}

func (op1 JoinDescriptor) IsRightAsscom(op2 JoinDescriptor) bool {
	if op1.operator.JoinType == join.JoinTypeInner || op1.operator.JoinType == join.JoinTypeCross {
		return op2.operator.JoinType == join.JoinTypeInner || op2.operator.JoinType == join.JoinTypeCross
	}

	if op1.operator.JoinType == join.JoinTypeFullOuter && op2.operator.JoinType == join.JoinTypeFullOuter {
		// TODO - footnote 4
		return op1.isNullRejecting && op2.isNullRejecting
	}

	return false
}
