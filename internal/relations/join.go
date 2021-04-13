package relations

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type joinRelation struct {
	left      Relation
	right     Relation
	condition expressions.BoolExpression
	fields    []shared.Field
}

var _ Relation = &joinRelation{}

func NewJoin(left Relation, right Relation, condition expressions.BoolExpression) Relation {
	return &joinRelation{
		left:      left,
		right:     right,
		condition: condition,
		fields:    append(joinFieldsForRelation(left), joinFieldsForRelation(right)...),
	}
}

func (r *joinRelation) Name() string           { return "" }
func (r *joinRelation) Fields() []shared.Field { return r.fields }

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

		if r.condition != nil {
			if ok, err := r.condition.ValueFrom(row); err != nil {
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
