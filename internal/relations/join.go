package relations

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type joinRelation struct {
	left   Relation
	right  Relation
	filter expressions.Expression
	fields []shared.Field
}

var _ Relation = &joinRelation{}

func NewJoin(left Relation, right Relation, condition expressions.Expression) Relation {
	return &joinRelation{
		left:   left,
		right:  right,
		filter: condition,
		fields: append(joinFieldsForRelation(left), joinFieldsForRelation(right)...),
	}
}

func (r *joinRelation) Name() string {
	return ""
}

func (r *joinRelation) Fields() []shared.Field {
	return copyFields(r.fields)
}

func (r *joinRelation) Serialize(w io.Writer, indentationLevel int) {
	indentation := indent(indentationLevel)
	io.WriteString(w, fmt.Sprintf("%sjoin\n", indentation))
	r.left.Serialize(w, indentationLevel+1)
	io.WriteString(w, fmt.Sprintf("%swith\n", indentation))
	r.right.Serialize(w, indentationLevel+1)

	if r.filter != nil {
		io.WriteString(w, fmt.Sprintf("%son %s\n", indentation, r.filter))
	}
}

func (r *joinRelation) Optimize() {
	if r.filter != nil {
		r.filter = r.distributeFilter(r.filter.Fold())
	}

	r.left.Optimize()
	r.right.Optimize()
}

func (r *joinRelation) distributeFilter(filter expressions.Expression) expressions.Expression {
	var conjunctions []expressions.Expression
	for _, expression := range filter.Conjunctions() {
		if !r.distributeExpression(expression) {
			conjunctions = append(conjunctions, expression)
		}
	}

	return combineConjunctions(conjunctions)
}

func (r *joinRelation) distributeExpression(expression expressions.Expression) bool {
	namesMatchingInLeft := false
	namesMissingFromLeft := false
	namesMatchingInRight := false
	namesMissingFromRight := false

	for _, field := range expression.Fields() {
		if _, err := shared.FindMatchingFieldIndex(field, r.left.Fields()); err != nil {
			namesMissingFromLeft = true
		} else {
			namesMatchingInLeft = true
		}
		if _, err := shared.FindMatchingFieldIndex(field, r.right.Fields()); err != nil {
			namesMissingFromRight = true
		} else {
			namesMatchingInRight = true
		}
	}

	pushedLeft := false
	if !namesMissingFromLeft {
		pushedLeft = r.left.PushDownFilter(expression)
	}

	pushedRight := false
	if !namesMissingFromRight {
		pushedRight = r.right.PushDownFilter(expression)
	}

	return (namesMatchingInLeft || !namesMissingFromLeft) == pushedLeft && (namesMatchingInRight || !namesMissingFromRight) == pushedRight
}

func (r *joinRelation) PushDownFilter(filter expressions.Expression) bool {
	if r.filter != nil {
		filter = expressions.NewAnd(r.filter, filter)
	}

	r.filter = filter
	return true
}

func (r *joinRelation) Scan(visitor VisitorFunc) error {
	return r.left.Scan(r.decorateLeftVisitor(visitor))
}

func (r *joinRelation) decorateLeftVisitor(visitor VisitorFunc) VisitorFunc {
	return func(leftRow shared.Row) (bool, error) {
		return true, r.right.Scan(r.decorateRightVisitor(visitor, leftRow))
	}
}

func (r *joinRelation) decorateRightVisitor(visitor VisitorFunc, leftRow shared.Row) VisitorFunc {
	return func(rightRow shared.Row) (bool, error) {
		row := shared.NewRow(r.Fields(), append(copyValues(leftRow.Values), rightRow.Values...))

		if r.filter != nil {
			if ok, err := shared.EnsureBool(r.filter.ValueFrom(row)); err != nil {
				return false, err
			} else if !ok {
				return true, nil
			}
		}

		return visitor(row)
	}
}

func joinFieldsForRelation(relation Relation) []shared.Field {
	if relation.Name() == "" {
		return relation.Fields()
	}

	return updateRelationName(relation.Fields(), relation.Name())
}
