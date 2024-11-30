package nodes

import (
	"fmt"
	"sort"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/ordering"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type logicalOrderNode struct {
	LogicalNode
	order impls.OrderExpression
}

var _ LogicalNode = &logicalOrderNode{}

func NewOrder(node LogicalNode, order impls.OrderExpression) LogicalNode {
	return &logicalOrderNode{
		LogicalNode: node,
		order:       order,
	}
}

func (n *logicalOrderNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {
	// We are nested in a parent sort and un-separated by an ordering boundary
	// (such as limit or offset). We'll ignore our old sort criteria and adopt
	// our parent since the ordering of rows at this point in the query should
	// not have an effect on the result.
	n.order = order
}

func (n *logicalOrderNode) Optimize(ctx impls.OptimizationContext) {
	if n.order != nil {
		n.order = n.order.Fold()
		n.LogicalNode.AddOrder(ctx, n.order)
	}

	n.LogicalNode.Optimize(ctx)

	if expressions.SubsumesOrder(n.order, n.LogicalNode.Ordering()) {
		n.order = nil
	}
}

func (n *logicalOrderNode) Ordering() impls.OrderExpression {
	if n.order == nil {
		return n.LogicalNode.Ordering()
	}

	return n.order
}

func (n *logicalOrderNode) SupportsMarkRestore() bool {
	return true
}

func (n *logicalOrderNode) Build() Node {
	if n.order == nil {
		return n.LogicalNode.Build()
	}

	return &orderNode{
		Node:   n.LogicalNode.Build(),
		order:  n.order,
		fields: n.Fields(),
	}
}

//
//

type orderNode struct {
	Node
	order  impls.OrderExpression
	fields []fields.Field
}

var _ Node = &orderNode{}

func (n *orderNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("order by %s", n.order)
	n.Node.Serialize(w.Indent())
}

func (n *orderNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Order Node scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	// TODO - commented out to support mark-restore
	// if n.order == nil {
	// 	return scanner, nil
	// }

	return newOrderScanner(ctx, scanner, n.fields, n.order)
}

type orderScanner struct {
	ctx     impls.ExecutionContext
	rows    rows.Rows
	indexes []int
	next    int
	mark    int
}

func newOrderScanner(ctx impls.ExecutionContext, scanner scan.RowScanner, fields []fields.Field, order impls.OrderExpression) (scan.RowScanner, error) {
	ctx.Log("Building Order scanner")

	rows, err := rows.NewRows(fields)
	if err != nil {
		return nil, err
	}

	rows, err = scan.ScanIntoRows(scanner, rows)
	if err != nil {
		return nil, err
	}

	indexes, err := findIndexIterationOrder(ctx, order, rows)
	if err != nil {
		return nil, err
	}

	return &orderScanner{
		ctx:     ctx,
		rows:    rows,
		indexes: indexes,
		mark:    -1,
	}, nil
}

func (s *orderScanner) Scan() (rows.Row, error) {
	s.ctx.Log("Scanning Order")

	if s.next < len(s.indexes) {
		row := s.rows.Row(s.indexes[s.next])
		s.next++
		return row, nil
	}

	return rows.Row{}, scan.ErrNoRows
}

func (s *orderScanner) Mark() {
	s.mark = s.next - 1
}

func (s *orderScanner) Restore() {
	if s.mark == -1 {
		panic("no mark to restore")
	}

	s.next = s.mark
}

func findIndexIterationOrder(ctx impls.ExecutionContext, order impls.OrderExpression, rows rows.Rows) ([]int, error) {
	expressions := order.Expressions()

	indexValues, err := makeIndexValues(ctx, expressions, rows)
	if err != nil {
		return nil, err
	}

	incomparable := false
	sort.SliceStable(indexValues, func(i, j int) bool {
		for k, value := range indexValues[i].values {
			reverse := expressions[k].Reverse

			switch ordering.CompareValues(value, indexValues[j].values[k]) {
			case ordering.OrderTypeIncomparable:
				incomparable = true
				return false
			case ordering.OrderTypeBefore:
				return !reverse
			case ordering.OrderTypeAfter:
				return reverse
			}
		}

		return false
	})
	if incomparable {
		return nil, fmt.Errorf("incomparable types")
	}

	indexes := make([]int, 0, len(indexValues))
	for _, value := range indexValues {
		indexes = append(indexes, value.index)
	}

	return indexes, nil
}

type indexValue struct {
	index  int
	values []any
}

func makeIndexValues(ctx impls.ExecutionContext, orderExpressions []impls.ExpressionWithDirection, rows rows.Rows) ([]indexValue, error) {
	indexValues := make([]indexValue, 0, len(rows.Values))
	for i := range rows.Values {
		var expressions []impls.Expression
		for _, expression := range orderExpressions {
			expressions = append(expressions, expression.Expression)
		}

		values, err := queries.EvaluateExpressions(ctx, expressions, rows.Row(i))
		if err != nil {
			return nil, err
		}

		indexValues = append(indexValues, indexValue{index: i, values: values})
	}

	return indexValues, nil
}
