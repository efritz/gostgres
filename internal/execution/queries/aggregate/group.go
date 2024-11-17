package aggregate

import (
	"strings"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/utils"
	"golang.org/x/exp/maps"
)

type hashAggregate struct {
	queries.Node
	groupExpressions []impls.Expression
	projection       *projection.Projection
}

var _ queries.Node = &hashAggregate{}

func NewHashAggregate(
	node queries.Node,
	groupExpressions []impls.Expression,
	projection *projection.Projection,
) queries.Node {
	return &hashAggregate{
		Node:             node,
		groupExpressions: groupExpressions,
		projection:       projection,
	}
}

func (n *hashAggregate) Name() string {
	return ""
}

func (n *hashAggregate) Fields() []fields.Field {
	return n.projection.Fields()
}

func (n *hashAggregate) Serialize(w serialization.IndentWriter) {
	var strExpressions []string
	for _, expr := range n.groupExpressions {
		strExpressions = append(strExpressions, expr.String())
	}

	w.WritefLine("group by %s, select(%s)", strings.Join(strExpressions, ", "), n.projection)
	n.Node.Serialize(w.Indent())
}

func (n *hashAggregate) AddFilter(filter impls.Expression) {
	n.Node.AddFilter(filter)
}

func (n *hashAggregate) AddOrder(order impls.OrderExpression) {
	// No-op
}

func (n *hashAggregate) Optimize() {
	n.Node.Optimize()
}

func (n *hashAggregate) Filter() impls.Expression {
	return n.Node.Filter()
}

func (n *hashAggregate) Ordering() impls.OrderExpression {
	return nil // TODO
}

func (n *hashAggregate) SupportsMarkRestore() bool {
	return false
}

func (n *hashAggregate) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Hash Aggregate scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	var exprs []impls.Expression
	for _, selectExpression := range n.projection.Aliases() {
		exprs = append(exprs, selectExpression.Expression)
	}

	h := map[uint64][]impls.AggregateExpression{}
	if err := scan.VisitRows(scanner, func(row rows.Row) (bool, error) {
		keys, err := queries.EvaluateExpressions(ctx, n.groupExpressions, row)
		if err != nil {
			return false, err
		}
		key := utils.Hash(keys)

		aggregateExpressions, ok := h[key]
		if !ok {
			for _, expression := range exprs {
				aggregate, err := expressions.AsAggregate(ctx, expression)
				if err != nil {
					return false, err
				}

				aggregateExpressions = append(aggregateExpressions, aggregate)
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

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Hash Aggregate")

		if i >= len(keys) {
			return rows.Row{}, scan.ErrNoRows
		}

		aggregateExpressions := h[keys[i]]
		i++

		var values []any
		for _, aggregateExpression := range aggregateExpressions {
			value, err := aggregateExpression.Done(ctx)
			if err != nil {
				return rows.Row{}, err
			}

			values = append(values, value)
		}

		return rows.Row{Fields: n.projection.Fields(), Values: values}, nil
	}), nil
}
