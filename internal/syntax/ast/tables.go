package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projector"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/access"
	"github.com/efritz/gostgres/internal/execution/queries/alias"
	"github.com/efritz/gostgres/internal/execution/queries/joins"
	"github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/ast/context"
)

type TargetTable struct {
	Name      string
	AliasName string
}

type TableReference struct {
	Name string

	table  impls.Table
	fields []fields.Field
}

func (r *TableReference) Resolve(ctx *context.ResolveContext) error {
	table, ok := ctx.Tables.Get(r.Name)
	if !ok {
		return fmt.Errorf("unknown table %q", r.Name)
	}

	tableFields := table.Fields()
	fields := make([]fields.Field, 0, len(tableFields))
	for _, tableField := range tableFields {
		fields = append(fields, tableField.Field)
	}

	// TODO - populate symbol table
	r.table = table
	r.fields = fields
	return nil
}

func (r *TableReference) TableFields() []fields.Field {
	return r.fields
}

func (r *TableReference) Build() (queries.Node, error) {
	return access.NewAccess(r.table), nil
}

type TableExpression struct {
	Base  AliasedTableReferenceOrExpression
	Joins []Join

	fields []fields.Field
}

type AliasedTableReferenceOrExpression struct {
	BaseTableExpression TableReferenceOrExpression
	Alias               *TableAlias
}

type TableReferenceOrExpression interface {
	BuilderResolver
	TableFields() []fields.Field
}

type TableAlias struct {
	TableAlias    string
	ColumnAliases []string
}

type Join struct {
	Table     *TableExpression
	Condition impls.Expression
}

func (e *TableExpression) Resolve(ctx *context.ResolveContext) error {
	if err := e.Base.BaseTableExpression.Resolve(ctx); err != nil {
		return err
	}

	for _, j := range e.Joins {
		if err := j.Table.Resolve(ctx); err != nil {
			return err
		}
	}

	baseFields := e.Base.BaseTableExpression.TableFields()

	var fields []fields.Field
	if a := e.Base.Alias; a != nil {
		for _, field := range baseFields {
			fields = append(fields, field.WithRelationName(a.TableAlias)) // TODO - do column aliases here too
		}
	} else {
		for _, field := range baseFields {
			fields = append(fields, field.WithRelationName(""))
		}
	}

	for _, j := range e.Joins {
		fields = append(fields, j.Table.TableFields()...)
	}

	// TODO - use symbol table
	e.fields = fields
	return nil
}

func (e TableExpression) TableFields() []fields.Field {
	return e.fields
}

func (e *TableExpression) Build() (queries.Node, error) {
	node, err := e.Base.BaseTableExpression.Build()
	if err != nil {
		return nil, err
	}

	if e.Base.Alias != nil {
		aliasName := e.Base.Alias.TableAlias
		columnNames := e.Base.Alias.ColumnAliases

		node = alias.NewAlias(node, aliasName)

		if len(columnNames) > 0 {
			var fields []fields.Field
			for _, f := range node.Fields() {
				if !f.Internal() {
					fields = append(fields, f)
				}
			}

			if len(columnNames) != len(fields) {
				return nil, fmt.Errorf("wrong number of fields in alias")
			}

			projectionExpressions := make([]projector.ProjectionExpression, 0, len(fields))
			for i, field := range fields {
				projectionExpressions = append(projectionExpressions, projector.NewAliasProjectionExpression(expressions.NewNamed(field), columnNames[i]))
			}

			node, err = projection.NewProjection(node, projectionExpressions)
			if err != nil {
				return nil, err
			}
		}
	}

	for _, j := range e.Joins {
		right, err := j.Table.Build()
		if err != nil {
			return nil, err
		}

		node = joins.NewJoin(node, right, j.Condition)
	}

	return node, nil
}

func joinNodes(left queries.Node, expressions []*TableExpression) queries.Node {
	for _, expression := range expressions {
		right, err := expression.Build()
		if err != nil {
			return nil
		}

		left = joins.NewJoin(left, right, nil)
	}

	return left
}
