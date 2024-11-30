package nodes

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
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

//
//

type EqualityPair struct {
	Left  impls.Expression
	Right impls.Expression
}

var LeftOfPair = func(pair EqualityPair) impls.Expression { return pair.Left }
var RightOfPair = func(pair EqualityPair) impls.Expression { return pair.Right }

func EvaluatePair(ctx impls.ExecutionContext, pairs []EqualityPair, expression func(EqualityPair) impls.Expression, row rows.Row) (values []any, _ error) {
	for _, pair := range pairs {
		value, err := queries.Evaluate(ctx, expression(pair), row)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}
