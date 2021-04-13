package relations

import "github.com/efritz/gostgres/internal/shared"

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

func (r *aliasRelation) Name() string           { return r.name }
func (r *aliasRelation) Fields() []shared.Field { return r.fields }
