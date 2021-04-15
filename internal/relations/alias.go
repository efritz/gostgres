package relations

import (
	"bytes"
	"fmt"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type aliasRelation struct {
	Relation
	name   string
	fields []shared.Field
}

var _ Relation = &aliasRelation{}

func NewAlias(relation Relation, name string) Relation {
	return &aliasRelation{
		Relation: relation,
		name:     name,
		fields:   updateRelationName(relation.Fields(), name),
	}
}

func (r *aliasRelation) Name() string {
	return r.name
}

func (r *aliasRelation) Fields() []shared.Field {
	return copyFields(r.fields)
}

func (r *aliasRelation) Serialize(buf *bytes.Buffer, indentationLevel int) {
	buf.WriteString(fmt.Sprintf("%salias as %s\n", indent(indentationLevel), r.name))
	r.Relation.Serialize(buf, indentationLevel+1)
}

func (r *aliasRelation) Optimize() {
	r.Relation.Optimize()
}

func (r *aliasRelation) PushDownFilter(filter expressions.Expression) bool {
	for _, field := range r.fields {
		filter = filter.Alias(field, expressions.NewNamed(shared.Field{
			RelationName: r.Relation.Name(),
			Name:         field.Name,
		}))
	}

	return r.Relation.PushDownFilter(filter)
}

func (r *aliasRelation) Scan(visitor VisitorFunc) error {
	return r.Relation.Scan(func(row shared.Row) (bool, error) {
		return visitor(shared.NewRow(r.fields, row.Values))
	})
}
