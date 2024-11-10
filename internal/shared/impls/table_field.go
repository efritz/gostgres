package impls

import (
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

type TableField struct {
	fields.ResolvedField
	nullable          bool
	defaultExpression Expression
}

func NewTableField(relationName, name string, typ types.Type, internal fields.InternalFieldType) TableField {
	descriptor := fields.NewFieldDescriptor(relationName, name, typ, internal)
	resolvedField := fields.ResolveDescriptor(descriptor)
	// 	return newTableField(resolvedField, true, nil)
	// }

	var (
		field                        = resolvedField
		nullable                     = true
		defaultExpression Expression = nil
	)

	// func newTableField(field fields.ResolvedField, nullable bool, defaultExpression Expression) TableField {
	return TableField{
		ResolvedField:     field,
		nullable:          nullable,
		defaultExpression: defaultExpression,
	}
}

func (f TableField) Nullable() bool {
	return f.nullable
}

func (f TableField) Default(ctx ExecutionContext) (any, error) {
	if f.defaultExpression == nil {
		return nil, nil
	}

	return f.defaultExpression.ValueFrom(ctx, rows.Row{})
}

// func (f TableField) WithRelationName(relationName string) TableField {
// 	return newTableField(f.Field.WithRelationName(relationName), f.nullable, f.defaultExpression)
// }

// func (f TableField) WithType(typ types.Type) TableField {
// 	return newTableField(f.Field.WithType(typ), f.nullable, f.defaultExpression)
// }

// TODO - make actual constraint?
// func (f TableField) WithNonNullable() TableField {
// 	return newTableField(f.Field, false, f.defaultExpression)
// }

// func (f TableField) WithDefault(defaultExpression Expression) TableField {
// 	return newTableField(f.Field, f.nullable, defaultExpression)
// }
