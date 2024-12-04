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

	fields                []fields.Field
	tableAliasProjection  *projection.Projection
	columnAliasProjection *projection.Projection
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

	baseFields := e.Base.BaseTableExpression.TableFields()

	if e.Base.Alias != nil {
		tableAlias := e.Base.Alias.TableAlias
		columnAliases := e.Base.Alias.ColumnAliases

		p, err := projection.NewProjectionFromProjectionExpressions(
			tableAlias,
			baseFields,
			[]projection.ProjectionExpression{
				projection.NewWildcardProjectionExpression(),
			},
		)
		if err != nil {
			return err
		}
		e.tableAliasProjection = p

		var rawFields []fields.Field
		for _, f := range baseFields {
			if !f.Internal() {
				rawFields = append(rawFields, f.WithRelationName(tableAlias))
			}
		}

		if len(columnAliases) > 0 {
			if len(columnAliases) != len(rawFields) {
				return fmt.Errorf("wrong number of fields in alias")
			}

			var projectionExpressions []projection.ProjectionExpression
			for i, field := range rawFields {
				alias := columnAliases[i]
				baseFields[i] = field.WithName(alias)
				projectionExpressions = append(projectionExpressions, projection.NewAliasedExpression(expressions.NewNamed(field), alias, false))
			}

			p, err := projection.NewProjectionFromProjectionExpressions(
				tableAlias,
				baseFields,
				projectionExpressions,
			)
			if err != nil {
				return err
			}
			e.columnAliasProjection = p
		} else {
			baseFields = rawFields
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

		resolved, err := ResolveExpression(ctx, j.Condition, nil, false)
		if err != nil {
			return err
		}
		j.Condition = resolved
	}

	e.fields = baseFields
	return nil
}

func (e *TableExpression) TableFields() []fields.Field {
	return slices.Clone(e.fields)
}

func (e *TableExpression) Build() (plan.LogicalNode, error) {
	node, err := e.Base.BaseTableExpression.Build()
	if err != nil {
		return nil, err
	}

	if e.tableAliasProjection != nil {
		node = plan.NewProjection(node, e.tableAliasProjection)
	}
	if e.columnAliasProjection != nil {
		node = plan.NewProjection(node, e.columnAliasProjection)
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
