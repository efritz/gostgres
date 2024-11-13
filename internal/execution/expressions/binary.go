package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

type binaryExpression struct {
	left         impls.Expression
	right        impls.Expression
	operatorText string
	typeChecker  binaryTypeChecker
	typ          types.Type
	valueFrom    binaryValueFromFunc
}

var _ impls.Expression = &binaryExpression{}

type binaryTypeChecker func(left types.Type, right types.Type) (types.Type, error)
type binaryValueFromFunc func(ctx impls.ExecutionContext, left, right impls.Expression, row rows.Row) (any, error)

func newBinaryExpression(left, right impls.Expression, operatorText string, typeChecker binaryTypeChecker, valueFrom binaryValueFromFunc) impls.Expression {
	return &binaryExpression{
		left:         left,
		right:        right,
		operatorText: operatorText,
		typeChecker:  typeChecker,
		valueFrom:    valueFrom,
	}
}

func (e binaryExpression) String() string {
	return fmt.Sprintf("%s %s %s", e.left, e.operatorText, e.right)
}

func (e *binaryExpression) Resolve(ctx impls.ExpressionResolutionContext) error {
	if err := e.left.Resolve(ctx); err != nil {
		return err
	}

	if err := e.right.Resolve(ctx); err != nil {
		return err
	}

	typ, err := e.typeChecker(e.left.Type(), e.right.Type())
	e.typ = typ
	return err
}

func (e binaryExpression) Type() types.Type {
	return e.typ
}

func (e binaryExpression) Equal(other impls.Expression) bool {
	o, ok := other.(*binaryExpression)
	if !ok {
		return false
	}

	if op, _, _ := IsArithmetic(&e); op != ArithmeticTypeUnknown {
		// Special case: if arithmetic operator is asymmetric, allow operands to be flipped
		if op.IsSymmetric() && e.operatorText == o.operatorText && e.left.Equal(o.right) && e.right.Equal(o.left) {
			return true
		}
	}

	if op, _, _ := IsComparison(&e); op != ComparisonTypeUnknown {
		// Special case: if comparison operators are inverted, allow operands to be flipped
		if o.operatorText == op.Flip().String() && e.left.Equal(o.right) && e.right.Equal(o.left) {
			return true
		}
	}

	// General case: equivalent operators and ordered operands are equal
	return e.operatorText == o.operatorText && e.left.Equal(o.left) && e.right.Equal(o.right)
}

func (e binaryExpression) Children() []impls.Expression {
	return []impls.Expression{e.left, e.right}
}

func (e binaryExpression) Fold() impls.Expression {
	return tryEvaluate(newBinaryExpression(e.left.Fold(), e.right.Fold(), e.operatorText, e.typeChecker, e.valueFrom))
}

func (e binaryExpression) Map(f func(impls.Expression) (impls.Expression, error)) (impls.Expression, error) {
	left, err := e.left.Map(f)
	if err != nil {
		return nil, err
	}

	right, err := e.right.Map(f)
	if err != nil {
		return nil, err
	}

	return f(newBinaryExpression(left, right, e.operatorText, e.typeChecker, e.valueFrom))
}

func (e binaryExpression) ValueFrom(ctx impls.ExecutionContext, row rows.Row) (any, error) {
	return e.valueFrom(ctx, e.left, e.right, row)
}
