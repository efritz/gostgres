package nodes

import (
	"fmt"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type btreeIndex struct {
	name        string
	table       *Table
	expressions []ExpressionWithDirection
	root        *btreeNode
}

type btreeNode struct {
	tid    int
	values []interface{}
	left   *btreeNode // TODO - make a proper BTree (not BST)
	right  *btreeNode
}

type btreeIndexScanOptions struct {
	lowerBound *scanBound
	upperBound *scanBound
}

type scanBound struct {
	expressions []expressions.Expression
	inclusive   bool
}

var _ Index[btreeIndexScanOptions] = &btreeIndex{}

func NewBTreeIndex(name string, table *Table, expressions []ExpressionWithDirection) *btreeIndex {
	return &btreeIndex{
		name:        name,
		table:       table,
		expressions: expressions,
	}
}

func (i *btreeIndex) Name() string {
	return i.name
}

func (i *btreeIndex) Filter() expressions.Expression {
	return nil
}

func (i *btreeIndex) Condition(opts btreeIndexScanOptions) expressions.Expression {
	var lowers []expressions.Expression
	if opts.lowerBound != nil {
		for j, expression := range i.expressions[:min(len(i.expressions), len(opts.lowerBound.expressions))] {
			if opts.lowerBound.expressions[j] != nil {
				if opts.lowerBound.inclusive {
					lowers = append(lowers, expressions.NewGreaterThanEquals(expression.Expression, opts.lowerBound.expressions[j]))
				} else {
					lowers = append(lowers, expressions.NewGreaterThan(expression.Expression, opts.lowerBound.expressions[j]))
				}
			}
		}
	}

	var uppers []expressions.Expression
	if opts.upperBound != nil {
		for j, expression := range i.expressions[:min(len(i.expressions), len(opts.upperBound.expressions))] {
			if opts.upperBound.expressions[j] != nil {
				if opts.upperBound.inclusive {
					uppers = append(uppers, expressions.NewLessThanEquals(expression.Expression, opts.upperBound.expressions[j]))
				} else {
					uppers = append(uppers, expressions.NewLessThan(expression.Expression, opts.upperBound.expressions[j]))
				}
			}
		}
	}

	// TODO - merge expressions like L <= b <= U where L = U to L = b

	var expr expressions.Expression
	for _, expression := range append(lowers, uppers...) {
		if expr == nil {
			expr = expression
		} else {
			expr = expressions.NewAnd(expr, expression)
		}
	}

	return expr
}

func (i *btreeIndex) Ordering() OrderExpression {
	return &orderExpression{expressions: i.expressions}
}

func (i *btreeIndex) Insert(row shared.Row) error {
	tid, values, err := i.extractTIDAndValuesFromRow(row)
	if err != nil {
		return err
	}

	i.root = i.root.insert(values, tid)
	return nil
}

func (n *btreeNode) insert(values []interface{}, tid int) *btreeNode {
	if n == nil {
		return &btreeNode{tid: tid, values: values}
	}

	switch shared.CompareValueSlices(values, n.values) {
	case shared.OrderTypeBefore, shared.OrderTypeEqual:
		n.left = n.left.insert(values, tid)
	case shared.OrderTypeAfter:
		n.right = n.right.insert(values, tid)
	}

	return n
}

func (i *btreeIndex) Delete(row shared.Row) error {
	tid, values, err := i.extractTIDAndValuesFromRow(row)
	if err != nil {
		return err
	}

	i.root = i.root.delete(values, tid)
	return nil
}

func (n *btreeNode) delete(values []interface{}, tid int) *btreeNode {
	if n == nil {
		return nil
	}

	if n.tid == tid {
		min := n.right
		for min.left != nil {
			min = min.left
		}

		n.values = min.values
		n.tid = min.tid
		n.right = n.right.delete(min.values, min.tid)
		return n
	}

	switch shared.CompareValueSlices(values, n.values) {
	case shared.OrderTypeBefore, shared.OrderTypeEqual:
		n.left = n.left.delete(values, tid)
	case shared.OrderTypeAfter:
		n.right = n.right.delete(values, tid)
	}

	return n
}

func (i *btreeIndex) Scanner(ctx ScanContext, opts btreeIndexScanOptions) (tidScanner, error) {
	stack := []*btreeNode{}
	current := i.root

	checkBound := func(nodeValues []interface{}, bound *scanBound, expected shared.OrderType) bool {
		if bound == nil {
			return true
		}

		var boundValues []interface{}
		for _, expression := range bound.expressions {
			value, err := ctx.Evaluate(expression, shared.Row{})
			if err != nil {
				return false
			}

			boundValues = append(boundValues, value)
		}

		orderType := shared.CompareValueSlices(nodeValues, boundValues)
		return orderType == expected || (orderType == shared.OrderTypeEqual && bound.inclusive)
	}

	withinLowerBound := func(values []interface{}) bool { return checkBound(values, opts.lowerBound, shared.OrderTypeAfter) }
	withinUpperBound := func(values []interface{}) bool { return checkBound(values, opts.upperBound, shared.OrderTypeBefore) }

	return tidScannerFunc(func() (int, error) {
		for current != nil || len(stack) > 0 {
			for current != nil {
				stack = append(stack, current)

				if withinLowerBound(current.values) {
					current = current.left
				} else {
					current = nil
				}
			}

			if len(stack) == 0 {
				break
			}

			idx := len(stack) - 1
			node := stack[idx]
			stack = stack[:idx]

			if withinUpperBound(node.values) {
				current = node.right

				if withinLowerBound(node.values) {
					return node.tid, nil
				}
			} else {
				current = nil
			}
		}

		return 0, ErrNoRows
	}), nil
}

func (i *btreeIndex) extractTIDAndValuesFromRow(row shared.Row) (int, []interface{}, error) {
	tid, ok := extractTID(row)
	if !ok {
		return 0, nil, fmt.Errorf("no tid in row")
	}

	values := []interface{}{}
	for _, expression := range i.expressions {
		value, err := expression.Expression.ValueFrom(row)
		if err != nil {
			return 0, nil, err
		}

		values = append(values, value)
	}

	return tid, values, nil
}
