package indexes

import (
	"fmt"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/table"
)

type btreeIndex struct {
	name        string
	tableName   string
	expressions []expressions.ExpressionWithDirection
	root        *btreeNode
}

type btreeNode struct {
	tid    int64
	values []any
	left   *btreeNode // TODO - make a proper BTree (not BST)
	right  *btreeNode
}

type BtreeIndexScanOptions struct {
	scanDirection ScanDirection
	lowerBounds   [][]scanBound
	upperBounds   [][]scanBound
}

type scanBound struct {
	expression expressions.Expression
	inclusive  bool
}

var _ Index[BtreeIndexScanOptions] = &btreeIndex{}

func NewBTreeIndex(name, tableName string, expressions []expressions.ExpressionWithDirection) *btreeIndex {
	return &btreeIndex{
		name:        name,
		tableName:   tableName,
		expressions: expressions,
	}
}

func (i *btreeIndex) Unwrap() table.Index {
	return i
}

func (i *btreeIndex) Filter() expressions.Expression {
	return nil
}

func (i *btreeIndex) Description(opts BtreeIndexScanOptions) string {
	direction := ""
	if opts.scanDirection == ScanDirectionBackward {
		direction = "backward "
	}

	return fmt.Sprintf("%sbtree index scan of %s via %s", direction, i.tableName, i.name)
}

func (i *btreeIndex) Condition(opts BtreeIndexScanOptions) expressions.Expression {
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

func (i *btreeIndex) conditionsForIndex(opts BtreeIndexScanOptions, index int) (lowers, uppers, equals []expressions.Expression) {
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

func (i *btreeIndex) Ordering(opts BtreeIndexScanOptions) expressions.OrderExpression {
	if opts.scanDirection == ScanDirectionBackward {
		var reversed []expressions.ExpressionWithDirection
		for _, expression := range i.expressions {
			reversed = append(reversed, expressions.ExpressionWithDirection{
				Expression: expression.Expression,
				Reverse:    !expression.Reverse,
			})
		}

		return expressions.NewOrderExpression(reversed)
	}

	return expressions.NewOrderExpression(i.expressions)
}

func (i *btreeIndex) Insert(row shared.Row) error {
	tid, values, err := i.extractTIDAndValuesFromRow(row)
	if err != nil {
		return err
	}

	i.root = i.root.insert(values, tid)
	return nil
}

func (n *btreeNode) insert(values []any, tid int64) *btreeNode {
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

func (n *btreeNode) delete(values []any, tid int64) *btreeNode {
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
