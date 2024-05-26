package aggregate

import (
	"fmt"
	"strings"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
	"golang.org/x/exp/maps"
)

type hashAggregate struct {
	queries.Node
	groupExpressions  []types.Expression
	selectExpressions []projection.ProjectionExpression
	projector         *projection.Projector
}

var _ queries.Node = &hashAggregate{}

func NewHashAggregate(
	node queries.Node,
	groupExpressions []types.Expression,
	selectExpressions []projection.ProjectionExpression,
) queries.Node {
	projector, err := projection.NewProjector(node.Name(), node.Fields(), selectExpressions)
	if err != nil {
		panic(err.Error()) // TODO
	}

	return &hashAggregate{
		Node:              node,
		groupExpressions:  groupExpressions,
		selectExpressions: selectExpressions,
		projector:         projector,
	}
}

func (n *hashAggregate) Name() string {
	return "" // TODO
}

func (n *hashAggregate) Fields() []shared.Field {
	return n.projector.Fields()
}

func (n *hashAggregate) Serialize(w serialization.IndentWriter) {
	var strExpressions []string
	for _, expr := range n.groupExpressions {
		strExpressions = append(strExpressions, expr.String())
	}

	w.WritefLine("group by %s, select(%s)", strings.Join(strExpressions, ", "), n.projector)
	n.Node.Serialize(w.Indent())
}

func (n *hashAggregate) AddFilter(filter types.Expression) {
	n.Node.AddFilter(filter)
}

func (n *hashAggregate) AddOrder(order expressions.OrderExpression) {
	// TODO
}

func (n *hashAggregate) Optimize() {
	n.Node.Optimize()
}

func (n *hashAggregate) Filter() types.Expression {
	return n.Node.Filter()
}

func (n *hashAggregate) Ordering() expressions.OrderExpression {
	return nil // TODO
}

func (n *hashAggregate) SupportsMarkRestore() bool {
	return false
}

func (n *hashAggregate) Scanner(ctx types.Context) (scan.Scanner, error) {
	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	var fields []shared.Field
	var exprs []types.Expression
	for _, selectExpression := range n.selectExpressions {
		expr, alias, ok := projection.UnwrapAlias(selectExpression)
		if !ok {
			return nil, fmt.Errorf("cannot unwrap alias %q", selectExpression)
		}

		fields = append(fields, shared.NewField("", alias, shared.TypeAny))
		exprs = append(exprs, expr)
	}

	h := map[uint64][]expressions.AggregateExpression{}
	if err := scan.VisitRows(scanner, func(row shared.Row) (bool, error) {
		keys, err := evaluatePair(ctx, n.groupExpressions, row)
		if err != nil {
			return false, err
		}
		key := shared.Hash(keys)

		aggregateExpressions, ok := h[key]
		if !ok {
			for _, expression := range exprs {
				aggregateExpressions = append(aggregateExpressions, expressions.AsAggregate(ctx, expression))
			}

			h[key] = aggregateExpressions
		}
		for _, aggregateExpression := range aggregateExpressions {
			if err := aggregateExpression.Step(ctx, row); err != nil {
				return false, err
			}
		}

		return true, nil
	}); err != nil {
		return nil, err
	}

	i := 0
	keys := maps.Keys(h)

	return scan.ScannerFunc(func() (shared.Row, error) {
		if i >= len(keys) {
			return shared.Row{}, scan.ErrNoRows
		}

		aggregateExpressions := h[keys[i]]
		i++

		var values []any
		for _, aggregateExpression := range aggregateExpressions {
			value, err := aggregateExpression.Done(ctx)
			if err != nil {
				return shared.Row{}, err
			}

			values = append(values, value)
		}

		return shared.Row{Fields: fields, Values: values}, nil
	}), nil
}

func evaluatePair(ctx types.Context, expressions []types.Expression, row shared.Row) (values []any, _ error) {
	for _, expression := range expressions {
		value, err := queries.Evaluate(ctx, expression, row)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}
