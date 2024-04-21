package shared

import "fmt"

type Field struct {
	relationName string
	name         string
	typ          Type
	internal     bool
}

func NewField(relationName, name string, typ Type) Field {
	return newField(relationName, name, typ, false)
}

func NewInternalField(relationName, name string, typ Type) Field {
	return newField(relationName, name, typ, true)
}

func newField(relationName, name string, typ Type, internal bool) Field {
	return Field{
		relationName: relationName,
		name:         name,
		typ:          typ,
		internal:     internal,
	}
}

func (f Field) RelationName() string { return f.relationName }
func (f Field) Name() string         { return f.name }
func (f Field) Type() Type           { return f.typ }
func (f Field) Internal() bool       { return f.internal }

func (f Field) String() string {
	if f.relationName == "" {
		return f.name
	}

	return fmt.Sprintf("%s.%s", f.relationName, f.name)
}

func (f Field) WithRelationName(relationName string) Field {
	return newField(relationName, f.name, f.typ, f.internal)
}

func (f Field) WithType(typ Type) Field {
	return newField(f.relationName, f.name, typ, f.internal)
}

func FindMatchingFieldIndex(needle Field, haystack []Field) (int, error) {
	unqualifiedIndexes := make([]int, 0, len(haystack))
	for index, field := range haystack {
		if field.name != needle.name {
			continue
		}

		if field.relationName == needle.relationName {
			return index, nil
		}

		unqualifiedIndexes = append(unqualifiedIndexes, index)
	}

	if needle.relationName == "" {
		if len(unqualifiedIndexes) == 1 {
			return unqualifiedIndexes[0], nil
		}
		if len(unqualifiedIndexes) > 1 {
			return 0, fmt.Errorf("ambiguous field %s", needle.Name())
		}
	}

	return 0, fmt.Errorf("unknown field %s", needle.Name())
}

func refineFieldTypes(fields []Field, values []any) ([]Field, error) {
	if len(fields) != len(values) {
		return nil, fmt.Errorf("unexpected number of columns")
	}

	refined := make([]Field, 0, len(fields))
	for i, field := range fields {
		refinedType, ok := field.typ.Refine(values[i])
		if !ok {
			fmt.Printf("> %#v\n", field)
			return nil, fmt.Errorf("type error (%v is not %s)", values[i], field.Type())
		}

		refined = append(refined, field.WithType(refinedType))
	}

	return refined, nil
}
