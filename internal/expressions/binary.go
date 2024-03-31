package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type binaryExpression struct {
	left         Expression
	right        Expression
	operatorText string
	valueFrom    binaryValueFromFunc
}

type binaryValueFromFunc func(left, right Expression, row shared.Row) (interface{}, error)

func newBinaryExpression(left, right Expression, operatorText string, valueFrom binaryValueFromFunc) Expression {
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

func (e binaryExpression) Equal(other Expression) bool {
	if o, ok := other.(binaryExpression); ok {
		// TODO - check for equivalence as well by possibly supplying a
		// symmetric operation and checking for lhs and rhs swapped:
		// e.operatorText == o.symmetricOperatorText && e.left.Equal(o.right) && e.right.Equal(o.left)
		//
		// This may not work for OR, though, as there is a tree equivalence
		// as well that we don't normalize. (a or b) or c != a or (b or c)
		// as the lhs and rhs are not equal. This may necessitate that we
		// somehow normalize all operators prior to comparison. This is also
		// an issue for addition and multiplication.
		//
		// A normalization path would mean that we only choose one set of operators
		// from an equivalency (rewrite > in terms of <), and we must also be able
		// to sort expressions in a deterministic way for operators that are associative.

		return e.operatorText == o.operatorText && e.left.Equal(o.left) && e.right.Equal(o.right)
	}

	return false
}

func (e binaryExpression) Fields() []shared.Field {
	return append(e.left.Fields(), e.right.Fields()...)
}

func (e binaryExpression) Named() (shared.Field, bool) {
	return shared.Field{}, false
}

func (e binaryExpression) Fold() Expression {
	return tryEvaluate(newBinaryExpression(e.left.Fold(), e.right.Fold(), e.operatorText, e.valueFrom))
}

func (e binaryExpression) Alias(field shared.Field, expression Expression) Expression {
	return newBinaryExpression(e.left.Alias(field, expression), e.right.Alias(field, expression), e.operatorText, e.valueFrom)
}

func (e binaryExpression) Conjunctions() []Expression {
	return []Expression{e}
}

func (e binaryExpression) ValueFrom(row shared.Row) (interface{}, error) {
	return e.valueFrom(e.left, e.right, row)
}
