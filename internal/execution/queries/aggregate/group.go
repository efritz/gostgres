package aggregate

import (
	"strings"

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
	aggregateFactory impls.AggregateExpressionFactory
}

var _ queries.Node = &hashAggregate{}

func NewHashAggregate(
	node queries.Node,
	groupExpressions []impls.Expression,
	projection *projection.Projection,
	aggregateFactory impls.AggregateExpressionFactory,
) queries.Node {
	return &hashAggregate{
		Node:             node,
		groupExpressions: groupExpressions,
		projection:       projection,
		aggregateFactory: aggregateFactory,
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
	return nil
}

func (n *hashAggregate) SupportsMarkRestore() bool {
	return false
}

func (n *hashAggregate) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Hash Aggregate scanner")

	buckets := map[uint64][]impls.AggregateExpression{}

	aggregatesForKey := func(key uint64) ([]impls.AggregateExpression, error) {
		aggregateExpressions, ok := buckets[key]
		if ok {
			return aggregateExpressions, nil
		}

		aggregateExpressions, err := n.aggregateFactory(ctx)
		if err != nil {
			return nil, err
		}

		buckets[key] = aggregateExpressions
		return aggregateExpressions, nil
	}

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	if err := scan.VisitRows(scanner, func(row rows.Row) (bool, error) {
		keys, err := queries.EvaluateExpressions(ctx, n.groupExpressions, row)
		if err != nil {
			return false, err
		}
		key := utils.Hash(keys)

		aggregateExpressions, err := aggregatesForKey(key)
		if err != nil {
			return false, err
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
	keys := maps.Keys(buckets)

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Hash Aggregate")

		if i >= len(keys) {
			return rows.Row{}, scan.ErrNoRows
		}

		aggregateExpressions := buckets[keys[i]]
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
