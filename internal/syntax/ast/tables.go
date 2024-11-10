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
	"github.com/efritz/gostgres/internal/shared/types"
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

func (r *TableReference) Resolve(ctx impls.ResolutionContext) error {
	table, ok := ctx.Catalog.Tables.Get(r.Name)
	if !ok {
		return fmt.Errorf("unknown table %q", r.Name)
	}

	tableFields := table.Fields()
	fields := make([]fields.Field, 0, len(tableFields))
	for _, tableField := range tableFields {
		fields = append(fields, tableField.Field)
	}

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

	fields                []fields.Field
	projectionExpressions []projector.ProjectionExpression
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

func (e *TableExpression) Resolve(ctx impls.ResolutionContext) error {
	ctx.PushScope()
	defer ctx.PopScope()

	if err := e.Base.BaseTableExpression.Resolve(ctx); err != nil {
		return err
	}

	baseFields := e.Base.BaseTableExpression.TableFields()

	var projectionExpressions []projector.ProjectionExpression
	if a := e.Base.Alias; a != nil {
		tableAlias := a.TableAlias
		columnAliases := a.ColumnAliases

		if len(columnAliases) > 0 {
			var rawFields []fields.Field
			for _, f := range baseFields {
				if !f.Internal() {
					rawFields = append(rawFields, f.WithRelationName(tableAlias))
				}
			}

			if len(columnAliases) != len(rawFields) {
				return fmt.Errorf("wrong number of fields in alias")
			}

			baseFields := baseFields[:0]
			for i, rawField := range rawFields {
				columnAlias := columnAliases[i]
				aliasedField := fields.NewField(tableAlias, columnAlias, types.TypeAny)

				baseFields = append(baseFields, aliasedField)
				projectionExpressions = append(projectionExpressions, projector.NewAliasProjectionExpression(expressions.NewNamed(rawField), columnAlias))
			}
		} else {
			for i, field := range baseFields {
				baseFields[i] = field.WithRelationName(tableAlias)
			}
		}
	}

	ctx.Bind(baseFields...)

	for _, j := range e.Joins {
		if err := j.Table.Resolve(ctx); err != nil {
			return err
		}

		joinFields := j.Table.TableFields()
		ctx.Bind(joinFields...)
		baseFields = append(baseFields, joinFields...)

		if j.Condition != nil {
			e, err := ctx.ResolveExpression(j.Condition)
			if err != nil {
				return err
			}
			j.Condition = e
		}
	}

	e.fields = baseFields
	e.projectionExpressions = projectionExpressions
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
		node = alias.NewAlias(node, e.Base.Alias.TableAlias)

		if len(e.projectionExpressions) > 0 {
			node, err = projection.NewProjection(node, e.projectionExpressions)
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
