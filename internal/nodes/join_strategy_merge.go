package nodes

import "github.com/efritz/gostgres/internal/expressions"

type mergeJoinStrategy struct {
	n              *joinNode
	left           expressions.Expression
	right          expressions.Expression
	comparisonType expressions.ComparisonType
}

func (s *mergeJoinStrategy) Name() string {
	return "merge"
}

func (s *mergeJoinStrategy) Ordering() OrderExpression {
	return nil // TODO - ordered on the left + right fields?
}

func (s *mergeJoinStrategy) Scanner() (Scanner, error) {
	// TODO - how to do this with scanners?
	panic("unimplemented")
}
