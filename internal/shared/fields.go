package shared

import "fmt"

type Field struct {
	RelationName string
	Name         string
	TypeKind     TypeKind
}

func NewField(relationName, name string, TypeKind TypeKind) Field {
	return Field{
		RelationName: relationName,
		Name:         name,
		TypeKind:     TypeKind,
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
