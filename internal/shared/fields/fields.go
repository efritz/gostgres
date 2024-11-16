package fields

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/types"
)

type InternalFieldType int

const (
	NonInternalField InternalFieldType = iota
	InternalFieldTid
)

type Field struct {
	relationName      string
	name              string
	typ               types.Type
	internalFieldType InternalFieldType
}

func NewField(relationName, name string, typ types.Type, internalFieldType InternalFieldType) Field {
	return newField(relationName, name, typ, internalFieldType)
}

func newField(relationName, name string, typ types.Type, internalFieldType InternalFieldType) Field {
	return Field{
		relationName:      relationName,
		name:              name,
		typ:               typ,
		internalFieldType: internalFieldType,
	}
}

func (f Field) RelationName() string { return f.relationName }
func (f Field) Name() string         { return f.name }
func (f Field) Type() types.Type     { return f.typ }
func (f Field) Internal() bool       { return f.internalFieldType != NonInternalField }
func (f Field) IsTID() bool          { return f.internalFieldType == InternalFieldTid }

func (f Field) String() string {
	if f.relationName == "" {
		return f.name
	}

	return fmt.Sprintf("%s.%s", f.relationName, f.name)
}
func (f Field) WithRelationName(relationName string) Field {
	return newField(relationName, f.name, f.typ, f.internalFieldType)
}

func (f Field) WithName(alias string) Field {
	return newField(f.relationName, alias, f.typ, f.internalFieldType)
}

func (f Field) WithType(typ types.Type) Field {
	return newField(f.relationName, f.name, typ, f.internalFieldType)
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
			return 0, fmt.Errorf("ambiguous field %q", needle.name)
		}
	}

	name := fmt.Sprintf("%q", needle.name)
	if needle.relationName != "" {
		name = fmt.Sprintf("%q.%s", needle.relationName, name)
	}

	return 0, fmt.Errorf("(runtime) unknown field %s", name)
}
