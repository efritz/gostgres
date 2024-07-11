package aggregate

import (
	"fmt"
	"strings"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projector"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/types"
	"github.com/efritz/gostgres/internal/shared/utils"
	"golang.org/x/exp/maps"
)

type hashAggregate struct {
	queries.Node
	groupExpressions  []impls.Expression
	selectExpressions []projector.ProjectionExpression
	projector         *projector.Projector
}

var _ queries.Node = &hashAggregate{}

func NewHashAggregate(
	node queries.Node,
	groupExpressions []impls.Expression,
	selectExpressions []projector.ProjectionExpression,
) queries.Node {
	projector, err := projector.NewProjector(node.Name(), node.Fields(), selectExpressions)
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

func (n *hashAggregate) Fields() []fields.Field {
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

func (n *hashAggregate) AddFilter(filter impls.Expression) {
	n.Node.AddFilter(filter)
}

func (n *hashAggregate) AddOrder(order impls.OrderExpression) {
	// TODO
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

func (n *hashAggregate) Scanner(ctx impls.Context) (scan.RowScanner, error) {
	ctx.Log("Building Hash Aggregate scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	var groupedFields []fields.Field
	var exprs []impls.Expression
	for _, selectExpression := range n.selectExpressions {
		expr, alias, ok := projector.UnwrapAlias(selectExpression)
		if !ok {
			return nil, fmt.Errorf("cannot unwrap alias %q", selectExpression)
		}

		groupedFields = append(groupedFields, fields.NewField("", alias, types.TypeAny))
		exprs = append(exprs, expr)
	}

	h := map[uint64][]impls.AggregateExpression{}
	if err := scan.VisitRows(scanner, func(row rows.Row) (bool, error) {
		keys, err := evaluatePair(ctx, n.groupExpressions, row)
		if err != nil {
			return false, err
		}
		key := utils.Hash(keys)

		aggregateExpressions, ok := h[key]
		if !ok {
			for _, expression := range exprs {
				a, _ := expressions.AsAggregate(ctx, expression)
				aggregateExpressions = append(aggregateExpressions, a)
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

		return rows.Row{Fields: groupedFields, Values: values}, nil
	}), nil
}

func evaluatePair(ctx impls.Context, expressions []impls.Expression, row rows.Row) (values []any, _ error) {
	for _, expression := range expressions {
		value, err := queries.Evaluate(ctx, expression, row)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}
