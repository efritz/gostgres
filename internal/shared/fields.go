package shared

import (
	"fmt"
)

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

type TableField struct {
	Field
	nullable     bool
	defaultValue func() any
}

func NewTableField(relationName, name string, typ Type) TableField {
	return newTableField(relationName, name, typ, false, true, nil)
}

func NewInternalTableField(relationName, name string, typ Type) TableField {
	return newTableField(relationName, name, typ, true, true, nil)
}

func newTableField(relationName, name string, typ Type, internal, nullable bool, defaultValue func() any) TableField {
	return TableField{
		Field:        newField(relationName, name, typ, internal),
		nullable:     nullable,
		defaultValue: defaultValue,
	}
}

func (f TableField) Nullable() bool {
	return f.nullable
}

func (f TableField) Default() (any, bool) {
	if f.defaultValue == nil {
		return nil, false
	}

	return f.defaultValue(), true
}

func (f TableField) WithRelationName(relationName string) TableField {
	return newTableField(relationName, f.name, f.typ, f.internal, f.nullable, f.defaultValue)
}

func (f TableField) WithType(typ Type) TableField {
	return newTableField(f.relationName, f.name, typ, f.internal, f.nullable, f.defaultValue)
}

// TODO - make actual constraint?
func (f TableField) WithNonNullable() TableField {
	return newTableField(f.relationName, f.name, f.typ, f.internal, false, f.defaultValue)
}

// TODO - stash expression
func (f TableField) WithDefault(defaultValue func() any) TableField {
	return newTableField(f.relationName, f.name, f.typ, f.internal, f.nullable, defaultValue)
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

	return 0, fmt.Errorf("unknown field %s", name)
}
