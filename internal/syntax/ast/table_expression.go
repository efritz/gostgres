package ast

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type TableExpression struct {
	Base  AliasedTableReferenceOrExpression
	Joins []Join

	fields     []fields.Field
	projection *projection.Projection
}

type TableReferenceOrExpression interface {
	BuilderResolver
	TableFields() []fields.Field
}

type AliasedTableReferenceOrExpression struct {
	BaseTableExpression TableReferenceOrExpression
	Alias               *TableAlias
}

type TableAlias struct {
	TableAlias    string
	ColumnAliases []string
}

type Join struct {
	Table     *TableExpression
	Condition impls.Expression
}

func (e *TableExpression) Resolve(ctx *impls.NodeResolutionContext) error {
	ctx.PushScope()
	defer ctx.PopScope()

	if err := e.Base.BaseTableExpression.Resolve(ctx); err != nil {
		return err
	}

	baseFields, aliasProjection, err := e.resolveTableAlias()
	if err != nil {
		return err
	}
	e.projection = aliasProjection

	ctx.Bind(baseFields...)

	for _, j := range e.Joins {
		if err := j.Table.Resolve(ctx); err != nil {
			return err
		}

		joinFields := j.Table.TableFields()
		ctx.Bind(joinFields...)
		baseFields = append(baseFields, joinFields...)

		resolved, err := ResolveExpression(ctx, j.Condition, nil, false)
		if err != nil {
			return err
		}
		j.Condition = resolved
	}

	e.fields = baseFields
	return nil
}

func (e *TableExpression) resolveTableAlias() ([]fields.Field, *projection.Projection, error) {
	if e.Base.Alias == nil {
		// No alias, return base fields without modification
		return e.Base.BaseTableExpression.TableFields(), nil, nil
	}

	tableAlias := e.Base.Alias.TableAlias
	columnAliases := e.Base.Alias.ColumnAliases
	baseFields := e.Base.BaseTableExpression.TableFields()

	var internalFields []fields.Field
	var nonInternalFields []fields.Field
	for _, field := range baseFields {
		if field.Internal() {
			internalFields = append(internalFields, field)
		} else {
			nonInternalFields = append(nonInternalFields, field)
		}
	}

	for _, field := range nonInternalFields[len(columnAliases):] {
		columnAliases = append(columnAliases, field.Name())
	}
	if len(columnAliases) > len(nonInternalFields) {
		return nil, nil, fmt.Errorf("has %d columns available but %d columns specified", len(nonInternalFields), len(columnAliases))
	}

	var projectionExpressions []projection.ProjectionExpression

	for i, field := range nonInternalFields {
		projectionExpressions = append(projectionExpressions, projection.NewAliasedExpression(
			expressions.NewNamed(field),
			columnAliases[i],
			false,
		))
	}

	for _, field := range internalFields {
		projectionExpressions = append(projectionExpressions, projection.NewAliasedExpression(
			expressions.NewNamed(field),
			field.Name(),
			true,
		))
	}

	p, err := projection.NewProjectionFromProjectionExpressions(tableAlias, baseFields, projectionExpressions, nil)
	if err != nil {
		return nil, nil, err
	}

	var fields []fields.Field
	for i, field := range nonInternalFields {
		fields = append(fields, field.WithRelationName(tableAlias).WithName(columnAliases[i]))
	}
	for _, field := range internalFields {
		fields = append(fields, field.WithRelationName(tableAlias))
	}

	return fields, p, nil
}

func (e *TableExpression) TableFields() []fields.Field {
	return slices.Clone(e.fields)
}

func (e *TableExpression) Build() (plan.LogicalNode, error) {
	node, err := e.Base.BaseTableExpression.Build()
	if err != nil {
		return nil, err
	}

	if e.projection != nil {
		node = plan.NewProjection(node, e.projection)
	}

	for _, j := range e.Joins {
		right, err := j.Table.Build()
		if err != nil {
			return nil, err
		}

		node = plan.NewJoin(node, right, j.Condition)
	}

	return node, nil
}
