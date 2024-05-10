package table

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type TableField struct {
	shared.Field
	nullable          bool
	defaultExpression expressions.Expression
}

func NewTableField(relationName, name string, typ shared.Type) TableField {
	return newTableField(shared.NewField(relationName, name, typ), true, nil)
}

func NewInternalTableField(relationName, name string, typ shared.Type) TableField {
	return newTableField(shared.NewInternalField(relationName, name, typ), true, nil)
}

func newTableField(field shared.Field, nullable bool, defaultExpression expressions.Expression) TableField {
	return TableField{
		Field:             field,
		nullable:          nullable,
		defaultExpression: defaultExpression,
	}
}

func (f TableField) Nullable() bool {
	return f.nullable
}

func (f TableField) Default(ctx expressions.ExpressionContext) (any, error) {
	if f.defaultExpression == nil {
		return nil, nil
	}

	return f.defaultExpression.ValueFrom(ctx, shared.Row{})
}

func (f TableField) WithRelationName(relationName string) TableField {
	return newTableField(f.Field.WithRelationName(relationName), f.nullable, f.defaultExpression)
}

func (f TableField) WithType(typ shared.Type) TableField {
	return newTableField(f.Field.WithType(typ), f.nullable, f.defaultExpression)
}

// TODO - make actual constraint?
func (f TableField) WithNonNullable() TableField {
	return newTableField(f.Field, false, f.defaultExpression)
}

func (f TableField) WithDefault(defaultExpression expressions.Expression) TableField {
	return newTableField(f.Field, f.nullable, defaultExpression)
}
