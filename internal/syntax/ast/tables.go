package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
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
	TableExpression(ctx BuildContext) (queries.Node, error)
}

type TableReferenceOrExpressionBuilder struct {
	TableReferenceOrExpression
}

func (r TableReferenceOrExpressionBuilder) Build(ctx BuildContext) (queries.Node, error) {
	return r.TableExpression(ctx)
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
	Name string
}

func (r TableReference) TableExpression(ctx BuildContext) (queries.Node, error) {
	table, ok := ctx.Tables.Get(r.Name)
	if !ok {
		return nil, fmt.Errorf("unknown table %s", r.Name)
	}

	return access.NewAccess(table), nil
}

type TableExpression struct {
	Base  AliasedTableReferenceOrExpression
	Joins []Join
}

type Join struct {
	Table     TableExpression
	Condition impls.Expression
}

func (e TableExpression) Build(ctx BuildContext) (queries.Node, error) {
	return e.TableExpression(ctx)
}

func (e TableExpression) TableExpression(ctx BuildContext) (queries.Node, error) {
	node, err := e.Base.BaseTableExpression.TableExpression(ctx)
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

			projectionExpressions := make([]projection.ProjectionExpression, 0, len(fields))
			for i, field := range fields {
				projectionExpressions = append(projectionExpressions, projection.NewAliasProjectionExpression(expressions.NewNamed(field), columnNames[i]))
			}

			node, err = projection.NewProjection(node, projectionExpressions)
			if err != nil {
				return nil, err
			}
		}
	}

	for _, j := range e.Joins {
		right, err := j.Table.TableExpression(ctx)
		if err != nil {
			return nil, err
		}

		node = joins.NewJoin(node, right, j.Condition)
	}

	return node, nil
}

func joinNodes(ctx BuildContext, left queries.Node, expressions []TableExpression) queries.Node {
	for _, expression := range expressions {
		right, err := expression.TableExpression(ctx)
		if err != nil {
			return nil
		}

		left = joins.NewJoin(left, right, nil)
	}

	return left
}
