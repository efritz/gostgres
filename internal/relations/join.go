package relations

import (
	"github.com/efritz/gostgres/internal/filters"
	"github.com/efritz/gostgres/internal/shared"
)

type joinRelation struct {
	left      Relation
	right     Relation
	condition filters.Filter
	fields    []shared.Field
}

var _ Relation = &joinRelation{}

func NewJoin(left Relation, right Relation, condition filters.Filter) Relation {
	return &joinRelation{
		left:      left,
		right:     right,
		condition: condition,
		fields:    append(joinFieldsForRelation(left), joinFieldsForRelation(right)...),
	}
}

func (r *joinRelation) Name() string           { return "" }
func (r *joinRelation) Fields() []shared.Field { return r.fields }

func (r *joinRelation) Scan(scanContext ScanContext, visitor VisitorFunc) error {
	return r.left.Scan(scanContext, r.decorateLeftVisitor(scanContext, visitor))
}

func (r *joinRelation) decorateLeftVisitor(scanContext ScanContext, visitor VisitorFunc) VisitorFunc {
	return func(scanContext ScanContext, leftValues []interface{}) (bool, error) {
		return true, r.right.Scan(scanContext, r.decorateRightVisitor(scanContext, visitor, leftValues))
	}
}

func (r *joinRelation) decorateRightVisitor(scanContext ScanContext, visitor VisitorFunc, leftValues []interface{}) VisitorFunc {
	fields := r.Fields()

	return func(scanContext ScanContext, rightValues []interface{}) (bool, error) {
		rowValues := append(copyValues(leftValues), rightValues...)

		if r.condition != nil {
			ok, err := r.condition.Test(shared.Row{Fields: fields, Values: rowValues})
			if err != nil {
				return false, err
			}

			if !ok {
				return true, nil
			}
		}

		return visitor(scanContext, rowValues)
	}
}

func joinFieldsForRelation(relation Relation) []shared.Field {
	if relation.Name() == "" {
		return relation.Fields()
	}

	return updateRelationName(relation.Fields(), relation.Name())
}
