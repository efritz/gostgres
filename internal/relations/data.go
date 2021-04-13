package relations

import "github.com/efritz/gostgres/internal/shared"

type dataRelation struct {
	name   string
	fields []shared.Field
	values [][]interface{}
}

var _ Relation = &dataRelation{}

func NewData(name string, fieldNames []string, values [][]interface{}) Relation {
	fields := make([]shared.Field, len(fieldNames))
	for i, field := range fieldNames {
		fields[i].Name = field
	}

	return &dataRelation{
		name:   name,
		fields: updateRelationName(fields, name),
		values: values,
	}
}

func (r *dataRelation) Name() string           { return r.name }
func (r *dataRelation) Fields() []shared.Field { return copyFields(r.fields) }

func (r *dataRelation) Scan(scanContext ScanContext, visitor VisitorFunc) error {
	for _, values := range r.values {
		if ok, err := visitor(scanContext, values); err != nil {
			return err
		} else if !ok {
			break
		}
	}

	return nil
}
