package nodes

import (
	"fmt"
	"sort"
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type OrderExpression interface {
	Fold() OrderExpression
	Expressions() []FieldExpression
}

type orderExpression struct {
	expressions []FieldExpression
}

func NewOrderExpression(expressions []FieldExpression) OrderExpression {
	return orderExpression{
		expressions: expressions,
	}
}

func (e orderExpression) String() string {
	parts := make([]string, 0, len(e.expressions))
	for _, expression := range e.expressions {
		part := fmt.Sprintf("%s", expression.Expression)
		if expression.Reverse {
			part += " desc"
		}

		parts = append(parts, part)
	}

	return strings.Join(parts, ", ")
}

func (e orderExpression) Fold() OrderExpression {
	expressions := make([]FieldExpression, 0, len(e.expressions))
	for _, expression := range e.expressions {
		expressions = append(expressions, expression.Fold())
	}

	return orderExpression{expressions: expressions}
}

func (e orderExpression) Expressions() []FieldExpression {
	return e.expressions
}

type FieldExpression struct {
	Expression expressions.Expression
	Reverse    bool
}

func (e FieldExpression) Fold() FieldExpression {
	return FieldExpression{
		Expression: e.Expression.Fold(),
		Reverse:    e.Reverse,
	}
}

func findIndexIterationOrder(order OrderExpression, rows shared.Rows) ([]int, error) {
	var expressions []FieldExpression
	if order != nil {
		expressions = order.Expressions()
	}

	indexValues, err := makeIndexValues(expressions, rows)
	if err != nil {
		return nil, err
	}

	incomparable := false
	sort.SliceStable(indexValues, func(i, j int) bool {
		for k, value := range indexValues[i].values {
			reverse := expressions[k].Reverse

			switch shared.CompareValues(value, indexValues[j].values[k]) {
			case shared.OrderTypeIncomparable:
				incomparable = true
				return false
			case shared.OrderTypeBefore:
				return !reverse
			case shared.OrderTypeAfter:
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
	values []interface{}
}

func makeIndexValues(expressions []FieldExpression, rows shared.Rows) ([]indexValue, error) {
	indexValues := make([]indexValue, 0, len(rows.Values))
	for i := range rows.Values {
		values := make([]interface{}, 0, len(expressions))
		for _, expression := range expressions {
			value, err := expression.Expression.ValueFrom(rows.Row(i))
			if err != nil {
				return nil, err
			}

			values = append(values, value)
		}

		indexValues = append(indexValues, indexValue{index: i, values: values})
	}

	return indexValues, nil
}
