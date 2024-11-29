package ast

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	projectionHelpers "github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/access"
	"github.com/efritz/gostgres/internal/execution/queries/joins"
	projection "github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type TargetTable struct {
	Name      string
	AliasName string
}

type TableReference struct {
	Name string

	table impls.Table
}

func (r *TableReference) Resolve(ctx *impls.NodeResolutionContext) error {
	table, ok := ctx.Catalog().Tables.Get(r.Name)
	if !ok {
		return fmt.Errorf("unknown table %q", r.Name)
	}
	r.table = table

	return nil
}

func (r *TableReference) TableFields() []fields.Field {
	var fields []fields.Field
	for _, f := range r.table.Fields() {
		fields = append(fields, f.Field)
	}

	return fields
}

func (r *TableReference) Build() (queries.Node, error) {
	return access.NewAccess(r.table), nil
}

type TableExpression struct {
	Base  AliasedTableReferenceOrExpression
	Joins []Join

	fields     []fields.Field
	projection *projectionHelpers.Projection
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

			var projectionExpressions []projectionHelpers.ProjectionExpression
			for i, field := range rawFields {
				alias := columnAliases[i]
				baseFields[i] = field.WithName(alias)
				projectionExpressions = append(projectionExpressions, projectionHelpers.NewAliasedExpression(expressions.NewNamed(field), alias))
			}

			p, err := projectionHelpers.NewProjectionFromProjectionExpressions(tableAlias, baseFields, projectionExpressions)
			if err != nil {
				return err
			}
			e.projection = p
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

		resolved, err := resolveExpression(ctx, j.Condition, nil, false)
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

func (e *TableExpression) Build() (queries.Node, error) {
	node, err := e.Base.BaseTableExpression.Build()
	if err != nil {
		return nil, err
	}

	if e.Base.Alias != nil {
		aliased, err := aliasTableName(node, e.Base.Alias.TableAlias)
		if err != nil {
			return nil, err
		}
		node = aliased

		if e.projection != nil {
			node = projection.NewProjection(node, e.projection)
		}
	}

	for _, j := range e.Joins {
		right, err := j.Table.Build()
		if err != nil {
			return nil, err
		}

		// TODO - We might want a SELECT/JOIN mega-node in which we can better control
		// the optimizatin of join order and strategy selection. This must include all
		// JOIN targets as well as the FROM.
		//
		// We may want the ability to do some amount of optimizations before we select
		// an order. We'll also want to do some more "global"-level rewrites, such as
		// pulling subqueries into the join list, that are tricky to do in tree form.
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
