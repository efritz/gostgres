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
)

type TargetTable struct {
	Name      string
	AliasName string
}

type TableReferenceOrExpression interface {
	TableExpression() // TODO - necessary marker?
	ResolverBuilder
}

type AliasedTableReferenceOrExpression struct {
	BaseTableExpression TableReferenceOrExpression
	Alias               *TableAlias
}

type TableAlias struct {
	TableAlias    string
	ColumnAliases []string
}

type TableReference struct {
	Name  string
	table impls.Table
}

func (r *TableReference) TableExpression() {}

func (r *TableReference) Resolve(ctx ResolveContext) ([]fields.Field, error) {
	table, ok := ctx.Tables.Get(r.Name)
	if !ok {
		return nil, fmt.Errorf("unknown table %q", r.Name)
	}
	r.table = table

	var fields []fields.Field
	for _, f := range table.Fields() {
		fields = append(fields, f.Field)
	}

	return fields, nil
}

func (r *TableReference) Build() (queries.Node, error) {
	return access.NewAccess(r.table), nil
}

type TableExpression struct {
	Base  AliasedTableReferenceOrExpression
	Joins []Join
}

type Join struct {
	Table     TableExpression
	Condition impls.Expression
}

func (e TableExpression) TableExpression() {}

func (e TableExpression) Resolve(ctx ResolveContext) ([]fields.Field, error) {
	fields, err := e.Base.BaseTableExpression.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	_ = e.Base.Alias // TODO

	for _, j := range e.Joins {
		fields2, err := j.Table.Resolve(ctx)
		if err != nil {
			return nil, err
		}

		fields = append(fields, fields2...)
	}

	return fields, nil
}

func (e TableExpression) Build() (queries.Node, error) {
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

func joinNodes(left queries.Node, expressions []TableExpression) queries.Node {
	for _, expression := range expressions {
		right, err := expression.Build()
		if err != nil {
			return nil
		}

		left = joins.NewJoin(left, right, nil)
	}

	return left
}
