package nodes

type mergeJoinStrategy struct {
	n *joinNode
}

func (s *mergeJoinStrategy) Name() string {
	return "merge"
}

func (s *mergeJoinStrategy) Ordering() OrderExpression {
	panic("merge join strategy unimplemented")
}

func (s *mergeJoinStrategy) Scan(visitor VisitorFunc) error {
	panic("merge join strategy unimplemented")
}
