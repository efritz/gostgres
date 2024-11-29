package combination

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/filter"
	"github.com/efritz/gostgres/internal/execution/queries/order"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/utils"
)

type logicalUnionNode struct {
	left     queries.LogicalNode
	right    queries.LogicalNode
	fields   []fields.Field
	distinct bool
}

var _ queries.LogicalNode = &logicalUnionNode{}

func NewUnion(left queries.LogicalNode, right queries.LogicalNode, distinct bool) (queries.LogicalNode, error) {
	leftFields := left.Fields()
	rightFields := right.Fields()

	if len(leftFields) != len(rightFields) {
		return nil, fmt.Errorf("unexpected union columns")
	}
	for i, leftField := range leftFields {
		if leftField.Type() != rightFields[i].Type() {
			// TODO - refine type if possible
			return nil, fmt.Errorf("unexpected union types")
		}
	}

	return &logicalUnionNode{
		left:     left,
		right:    right,
		fields:   leftFields,
		distinct: distinct,
	}, nil
}

func (n *logicalUnionNode) Name() string {
	return ""
}

func (n *logicalUnionNode) Fields() []fields.Field {
	return slices.Clone(n.fields)
}

func (n *logicalUnionNode) AddFilter(ctx impls.OptimizationContext, filterExpression impls.Expression) {
	filter.LowerFilter(ctx, filterExpression, n.left, n.right)
}

func (n *logicalUnionNode) AddOrder(ctx impls.OptimizationContext, orderExpression impls.OrderExpression) {
	order.LowerOrder(ctx, orderExpression, n.left, n.right)
}

func (n *logicalUnionNode) Optimize(ctx impls.OptimizationContext) {
	n.left.Optimize(ctx)
	n.right.Optimize(ctx)
}

func (n *logicalUnionNode) Filter() impls.Expression {
	return expressions.FilterIntersection(n.left.Filter(), n.right.Filter())
}

func (n *logicalUnionNode) Ordering() impls.OrderExpression { return nil }
func (n *logicalUnionNode) SupportsMarkRestore() bool       { return false }

func (n *logicalUnionNode) Build() queries.Node {
	return &unionNode{
		left:     n.left.Build(),
		right:    n.right.Build(),
		fields:   n.fields,
		distinct: n.distinct,
	}
}

//
//

type unionNode struct {
	left     queries.Node
	right    queries.Node
	fields   []fields.Field
	distinct bool
}

var _ queries.Node = &unionNode{}

func (n *unionNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("union")
	n.left.Serialize(w.Indent())
	w.WritefLine("with")
	n.right.Serialize(w.Indent())
}

func (n *unionNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Union scanner")

	hash := map[string]struct{}{}
	mark := func(row rows.Row) bool {
		key := utils.HashSlice(row.Values)
		if _, ok := hash[key]; ok {
			return !n.distinct
		}

		hash[key] = struct{}{}
		return true
	}

	leftScanner, err := n.left.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	var rightScanner scan.RowScanner

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Union")

		for leftScanner != nil {
			row, err := leftScanner.Scan()
			if err != nil {
				if err == scan.ErrNoRows {
					leftScanner = nil
					continue
				}

				return rows.Row{}, err
			}

			if mark(row) {
				return row, nil
			}
		}

		if rightScanner == nil {
			rightScanner, err = n.right.Scanner(ctx)
			if err != nil {
				return rows.Row{}, err
			}
		}

		for {
			row, err := rightScanner.Scan()
			if err != nil {
				return rows.Row{}, err
			}

			if mark(row) {
				return rows.NewRow(n.fields, row.Values)
			}
		}
	}), nil
}
