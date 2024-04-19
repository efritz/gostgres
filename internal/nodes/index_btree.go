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
	values []any
	left   *btreeNode // TODO - make a proper BTree (not BST)
	right  *btreeNode
}

type btreeIndexScanOptions struct {
	scanDirection ScanDirection
	lowerBounds   [][]scanBound
	upperBounds   [][]scanBound
}

type ScanDirection int

const (
	ScanDirectionUnknown ScanDirection = iota
	ScanDirectionForward
	ScanDirectionBackward
)

type scanBound struct {
	expression expressions.Expression
	inclusive  bool
}

var _ Index[btreeIndexScanOptions] = &btreeIndex{}

func NewBTreeIndex(name string, table *Table, expressions []ExpressionWithDirection) *btreeIndex {
	return &btreeIndex{
		name:        name,
		table:       table,
		expressions: expressions,
	}
}

func (i *btreeIndex) Unwrap() BaseIndex {
	return i
}

func (i *btreeIndex) Filter() expressions.Expression {
	return nil
}

func (i *btreeIndex) Description(opts btreeIndexScanOptions) string {
	direction := ""
	if opts.scanDirection == ScanDirectionBackward {
		direction = "backward "
	}

	return fmt.Sprintf("%sbtree index scan of %s via %s", direction, i.table.name, i.name)
}

func (i *btreeIndex) Condition(opts btreeIndexScanOptions) expressions.Expression {
	var allExpressions []expressions.Expression

	for j := range i.expressions {
		lowers, uppers, equals := i.conditionsForIndex(opts, j)

		allExpressions = append(allExpressions, lowers...)
		allExpressions = append(allExpressions, uppers...)
		allExpressions = append(allExpressions, equals...)
	}

	var expr expressions.Expression
	for _, expression := range allExpressions {
		if expr == nil {
			expr = expression
		} else {
			expr = expressions.NewAnd(expr, expression)
		}
	}

	return expr
}

func (i *btreeIndex) conditionsForIndex(opts btreeIndexScanOptions, index int) (lowers, uppers, equals []expressions.Expression) {
	var lowerBounds []scanBound
	if index < len(opts.lowerBounds) {
		lowerBounds = opts.lowerBounds[index]
	}

	var upperBounds []scanBound
	if index < len(opts.upperBounds) {
		upperBounds = opts.upperBounds[index]
	}

	skipLowers := map[int]struct{}{}
	skipUppers := map[int]struct{}{}
	expression := i.expressions[index].Expression

	for j, lowerBound := range lowerBounds {
		if !lowerBound.inclusive {
			continue
		}

		for k, upperBound := range upperBounds {
			if upperBound.inclusive && lowerBound.expression.Equal(upperBound.expression) {
				skipLowers[j] = struct{}{}
				skipUppers[k] = struct{}{}

				// TODO - should apply this more generally in conjunction folding
				equals = append(equals, expressions.NewEquals(expression, lowerBound.expression))
			}
		}
	}

	for _, lowerBound := range lowerBounds {
		if _, ok := skipLowers[index]; ok {
			continue
		}

		if lowerBound.inclusive {
			lowers = append(lowers, expressions.NewGreaterThanEquals(expression, lowerBound.expression))
		} else {
			lowers = append(lowers, expressions.NewGreaterThan(expression, lowerBound.expression))
		}
	}

	for _, upperBound := range upperBounds {
		if _, ok := skipUppers[index]; ok {
			continue
		}

		if upperBound.inclusive {
			uppers = append(uppers, expressions.NewLessThanEquals(expression, upperBound.expression))
		} else {
			uppers = append(uppers, expressions.NewLessThan(expression, upperBound.expression))
		}
	}

	return lowers, uppers, equals
}

func (i *btreeIndex) Ordering(opts btreeIndexScanOptions) OrderExpression {
	if opts.scanDirection == ScanDirectionBackward {
		var reversed []ExpressionWithDirection
		for _, expression := range i.expressions {
			reversed = append(reversed, ExpressionWithDirection{
				Expression: expression.Expression,
				Reverse:    !expression.Reverse,
			})
		}

		return &orderExpression{expressions: reversed}
	}

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

func (n *btreeNode) insert(values []any, tid int) *btreeNode {
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

func (n *btreeNode) delete(values []any, tid int) *btreeNode {
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

	checkBounds := func(nodeValues []any, bounds [][]scanBound, expected shared.OrderType) bool {
		for j, bounds := range bounds[:min(len(bounds), len(nodeValues))] {
			for _, bound := range bounds {
				value, err := ctx.Evaluate(bound.expression, shared.Row{})
				if err != nil {
					return false
				}

				orderType := shared.CompareValues(nodeValues[j], value)
				if !(orderType == expected || (orderType == shared.OrderTypeEqual && bound.inclusive)) {
					return false
				}
			}
		}

		return true
	}

	withinLowerBound := func(values []any) bool { return checkBounds(values, opts.lowerBounds, shared.OrderTypeAfter) }
	withinUpperBound := func(values []any) bool { return checkBounds(values, opts.upperBounds, shared.OrderTypeBefore) }

	return tidScannerFunc(func() (int, error) {
		for current != nil || len(stack) > 0 {
			for current != nil {
				stack = append(stack, current)

				if opts.scanDirection != ScanDirectionBackward && withinLowerBound(current.values) {
					current = current.left
				} else if opts.scanDirection == ScanDirectionBackward == withinUpperBound(current.values) {
					current = current.right
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

			if opts.scanDirection == ScanDirectionForward && withinUpperBound(node.values) {
				current = node.right
			} else if opts.scanDirection == ScanDirectionBackward && withinLowerBound(node.values) {
				current = node.left
			} else {
				current = nil
			}

			if withinLowerBound(node.values) && withinUpperBound(node.values) {
				return node.tid, nil
			}
		}

		return 0, ErrNoRows
	}), nil
}

func (i *btreeIndex) extractTIDAndValuesFromRow(row shared.Row) (int, []any, error) {
	tid, ok := extractTID(row)
	if !ok {
		return 0, nil, fmt.Errorf("no tid in row")
	}

	values := []any{}
	for _, expression := range i.expressions {
		value, err := expression.Expression.ValueFrom(row)
		if err != nil {
			return 0, nil, err
		}

		values = append(values, value)
	}

	return tid, values, nil
}
