package shared

import "fmt"

type Field struct {
	RelationName string
	Name         string
	TypeKind     TypeKind
	Internal     bool
}

func NewField(relationName, name string, TypeKind TypeKind, internal bool) Field {
	return Field{
		RelationName: relationName,
		Name:         name,
		TypeKind:     TypeKind,
		Internal:     internal,
	}
}

func FindMatchingFieldIndex(needle Field, haystack []Field) (int, error) {
	unqualifiedIndexes := make([]int, 0, len(haystack))
	for index, field := range haystack {
		if field.Name != needle.Name {
			continue
		}

		if field.RelationName == needle.RelationName {
			return index, nil
		}

		unqualifiedIndexes = append(unqualifiedIndexes, index)
	}

	if needle.RelationName == "" {
		if len(unqualifiedIndexes) == 1 {
			return unqualifiedIndexes[0], nil
		}
		if len(unqualifiedIndexes) > 1 {
			return 0, fmt.Errorf("ambiguous field %s", needle.Name)
		}
	}

	return 0, fmt.Errorf("unknown field %s", needle.Name)
}

func refineFieldTypes(fields []Field, values []interface{}) ([]Field, error) {
	if len(fields) != len(values) {
		return nil, fmt.Errorf("unexpected number of columns")
	}

	refined := make([]Field, 0, len(fields))
	for i, field := range fields {
		refinedType := refineType(field.TypeKind, values[i])
		if refinedType == TypeKindInvalid {
			return nil, fmt.Errorf("type error (%v is not %s)", values[i], field.TypeKind)
		}

		refined = append(refined, NewField(field.RelationName, field.Name, refinedType, field.Internal))
	}

	return refined, nil
}
