package indexes

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/ordering"
	"github.com/efritz/gostgres/internal/shared/rows"
)

type btreeIndex struct {
	name        string
	tableName   string
	unique      bool
	expressions []impls.ExpressionWithDirection
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
	expression impls.Expression
	inclusive  bool
}

func NewBtreeSearchOptions(values []any) BtreeIndexScanOptions {
	var lowerBounds [][]scanBound
	var upperBounds [][]scanBound

	for _, value := range values {
		bound := []scanBound{{
			expression: expressions.NewConstant(value),
			inclusive:  true,
		}}
		lowerBounds = append(lowerBounds, bound)
		upperBounds = append(upperBounds, bound)
	}

	return BtreeIndexScanOptions{
		scanDirection: ScanDirectionForward,
		lowerBounds:   lowerBounds,
		upperBounds:   upperBounds,
	}
}

var _ impls.Index[BtreeIndexScanOptions] = &btreeIndex{}

func NewBTreeIndex(name, tableName string, unique bool, expressions []impls.ExpressionWithDirection) impls.Index[BtreeIndexScanOptions] {
	return &btreeIndex{
		name:        name,
		tableName:   tableName,
		unique:      unique,
		expressions: expressions,
	}
}

func (i *btreeIndex) Unwrap() impls.BaseIndex {
	return i
}

func (i *btreeIndex) UniqueOn() []fields.Field {
	if !i.unique {
		return nil
	}

	var fields []fields.Field
	for _, e := range i.expressions {
		named, ok := e.Expression.(expressions.NamedExpression)
		if !ok {
			return nil
		}

		fields = append(fields, named.Field())
	}

	return fields
}

func (i *btreeIndex) Filter() impls.Expression {
	return nil
}

func (i *btreeIndex) Name() string {
	return i.name
}

func (i *btreeIndex) Description(opts BtreeIndexScanOptions) string {
	direction := ""
	if opts.scanDirection == ScanDirectionBackward {
		direction = "backward "
	}

	return fmt.Sprintf("%sbtree index scan of %s via %s", direction, i.tableName, i.name)
}

func (i *btreeIndex) Condition(opts BtreeIndexScanOptions) impls.Expression {
	var allExpressions []impls.Expression

	for j := range i.expressions {
		lowers, uppers, equals := i.conditionsForIndex(opts, j)

		allExpressions = append(allExpressions, lowers...)
		allExpressions = append(allExpressions, uppers...)
		allExpressions = append(allExpressions, equals...)
	}

	var expr impls.Expression
	for _, expression := range allExpressions {
		if expr == nil {
			expr = expression
		} else {
			expr = expressions.NewAnd(expr, expression)
		}
	}

	return expr
}

func (i *btreeIndex) conditionsForIndex(opts BtreeIndexScanOptions, index int) (lowers, uppers, equals []impls.Expression) {
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

func (i *btreeIndex) Ordering(opts BtreeIndexScanOptions) impls.OrderExpression {
	if opts.scanDirection == ScanDirectionBackward {
		var reversed []impls.ExpressionWithDirection
		for _, expression := range i.expressions {
			reversed = append(reversed, impls.ExpressionWithDirection{
				Expression: expression.Expression,
				Reverse:    !expression.Reverse,
			})
		}

		return expressions.NewOrderExpression(reversed)
	}

	return expressions.NewOrderExpression(i.expressions)
}

func (i *btreeIndex) Insert(row rows.Row) error {
	tid, values, err := i.extractTIDAndValuesFromRow(row)
	if err != nil {
		return err
	}

	i.root, err = i.root.insert(values, tid, i.unique)
	return err
}

func (n *btreeNode) insert(values []any, tid int64, unique bool) (*btreeNode, error) {
	if n == nil {
		return &btreeNode{tid: tid, values: values}, nil
	}

	switch ordering.CompareValueSlices(values, n.values) {
	case ordering.OrderTypeEqual:
		if unique {
			return nil, fmt.Errorf("unique constraint violation")
		}
		fallthrough

	case ordering.OrderTypeBefore:
		newLeft, err := n.left.insert(values, tid, unique)
		if err != nil {
			return nil, err
		}
		n.left = newLeft

	case ordering.OrderTypeAfter:
		newRight, err := n.right.insert(values, tid, unique)
		if err != nil {
			return nil, err
		}
		n.right = newRight
	}

	return n, nil
}

func (i *btreeIndex) Delete(row rows.Row) error {
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
		if n.right == nil {
			return n.left
		}

		min := n.right
		for min.left != nil {
			min = min.left
		}

		n.values = min.values
		n.tid = min.tid
		n.right = n.right.delete(min.values, min.tid)
		return n
	}

	switch ordering.CompareValueSlices(values, n.values) {
	case ordering.OrderTypeBefore, ordering.OrderTypeEqual:
		n.left = n.left.delete(values, tid)
	case ordering.OrderTypeAfter:
		n.right = n.right.delete(values, tid)
	}

	return n
}

func (i *btreeIndex) extractTIDAndValuesFromRow(row rows.Row) (int64, []any, error) {
	tid, err := row.TID()
	if err != nil {
		return 0, nil, err
	}

	values := []any{}
	for _, expression := range i.expressions {
		value, err := expression.Expression.ValueFrom(impls.EmptyExecutionContext, row)
		if err != nil {
			return 0, nil, err
		}

		values = append(values, value)
	}

	return tid, values, nil
}
