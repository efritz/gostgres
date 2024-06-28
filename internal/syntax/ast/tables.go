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

// TODO - can we get rid of this iface?
type TableReferenceOrExpression interface {
	TableExpression()
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
	Name string
}

func (r TableReference) Resolve(ctx ResolveContext) error {
	return fmt.Errorf("table reference resolve unimplemented")
}

func (r TableReference) TableExpression() {}

func (r TableReference) Build(ctx BuildContext) (queries.Node, error) {
	table, ok := ctx.Tables.Get(r.Name)
	if !ok {
		return nil, fmt.Errorf("unknown table %q", r.Name)
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

func (e TableExpression) Resolve(ctx ResolveContext) error {
	return fmt.Errorf("table expression resolve unimplemented")
}

func (e TableExpression) TableExpression() {}

func (e TableExpression) Build(ctx BuildContext) (queries.Node, error) {
	node, err := e.Base.BaseTableExpression.Build(ctx)
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
		right, err := j.Table.Build(ctx)
		if err != nil {
			return nil, err
		}

		node = joins.NewJoin(node, right, j.Condition)
	}

	return node, nil
}

func joinNodes(ctx BuildContext, left queries.Node, expressions []TableExpression) queries.Node {
	for _, expression := range expressions {
		right, err := expression.Build(ctx)
		if err != nil {
			return nil
		}

		left = joins.NewJoin(left, right, nil)
	}

	return left
}
