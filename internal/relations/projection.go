package relations

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type projectionRelation struct {
	Relation
	projector *projector
}

var _ Relation = &projectionRelation{}

func NewProjection(relation Relation, expressions []ProjectionExpression) (Relation, error) {
	projector, err := newProjector(relation.Name(), relation.Fields(), expressions)
	if err != nil {
		return nil, err
	}

	return &projectionRelation{
		Relation:  relation,
		projector: projector,
	}, nil
}

func (r *projectionRelation) Fields() []shared.Field {
	return copyFields(r.projector.fields)
}

func (r *projectionRelation) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%sselect (%s)\n", indent(indentationLevel), r.projector))
	r.Relation.Serialize(w, indentationLevel+1)
}

func (r *projectionRelation) Optimize() {
	r.projector.optimize()
	r.Relation.Optimize()
}

func (r *projectionRelation) PushDownFilter(filter expressions.Expression) bool {
	return r.Relation.PushDownFilter(r.projector.projectExpression(filter))
}

func (r *projectionRelation) Scan(visitor VisitorFunc) error {
	return r.Relation.Scan(r.decorateVisitor(visitor))
}

func (r *projectionRelation) decorateVisitor(visitor VisitorFunc) VisitorFunc {
	return func(row shared.Row) (bool, error) {
		row, err := r.projector.projectRow(row)
		if err != nil {
			return false, err
		}

		return visitor(row)
	}
}
