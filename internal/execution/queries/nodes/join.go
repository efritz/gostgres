package nodes

import (
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type JoinStrategy interface {
	Name() string
	Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error)
}

type joinNode struct {
	left     Node
	right    Node
	filter   impls.Expression
	fields   []fields.Field
	strategy JoinStrategy
}

func NewJoin(left, right Node, filter impls.Expression, fields []fields.Field, strategy JoinStrategy) Node {
	return &joinNode{
		left:     left,
		right:    right,
		filter:   filter,
		fields:   fields,
		strategy: strategy,
	}
}

func (n *joinNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("join using %s", n.strategy.Name())
	n.left.Serialize(w.Indent())
	w.WritefLine("with")
	n.right.Serialize(w.Indent())

	if n.filter != nil {
		w.WritefLine("on %s", n.filter)
	}
}

func (n *joinNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	if n.strategy == nil {
		panic("No strategy set - optimization required before scanning can be performed")
	}

	return n.strategy.Scanner(ctx)
}
