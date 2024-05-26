package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

type binaryExpression struct {
	left         types.Expression
	right        types.Expression
	operatorText string
	valueFrom    binaryValueFromFunc
}

type binaryValueFromFunc func(ctx types.Context, left, right types.Expression, row shared.Row) (any, error)

func newBinaryExpression(left, right types.Expression, operatorText string, valueFrom binaryValueFromFunc) types.Expression {
	return binaryExpression{
		left:         left,
		right:        right,
		operatorText: operatorText,
		valueFrom:    valueFrom,
	}
}

func (e binaryExpression) String() string {
	return fmt.Sprintf("%s %s %s", e.left, e.operatorText, e.right)
}

func (e binaryExpression) Equal(other types.Expression) bool {
	o, ok := other.(binaryExpression)
	if !ok {
		return false
	}

	if op, _, _ := IsArithmetic(e); op != ArithmeticTypeUnknown {
		// Special case: if arithmetic operator is asymmetric, allow operands to be flipped
		if op.IsSymmetric() && e.operatorText == o.operatorText && e.left.Equal(o.right) && e.right.Equal(o.left) {
			return true
		}
	}

	if op, _, _ := IsComparison(e); op != ComparisonTypeUnknown {
		// Special case: if comparison operators are inverted, allow operands to be flipped
		if o.operatorText == op.Flip().String() && e.left.Equal(o.right) && e.right.Equal(o.left) {
			return true
		}
	}

	// General case: equivalent operators and ordered operands are equal
	return e.operatorText == o.operatorText && e.left.Equal(o.left) && e.right.Equal(o.right)
}

func (e binaryExpression) Children() []types.Expression {
	return []types.Expression{e.left, e.right}
}

func (e binaryExpression) Fold() types.Expression {
	return tryEvaluate(newBinaryExpression(e.left.Fold(), e.right.Fold(), e.operatorText, e.valueFrom))
}

func (e binaryExpression) Map(f func(types.Expression) types.Expression) types.Expression {
	return f(newBinaryExpression(e.left.Map(f), e.right.Map(f), e.operatorText, e.valueFrom))
}

func (e binaryExpression) ValueFrom(ctx types.Context, row shared.Row) (any, error) {
	return e.valueFrom(ctx, e.left, e.right, row)
}
