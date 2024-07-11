package ast

import (
	"fmt"
	"strings"

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

type TableReferenceOrExpression interface {
	ResolverBuilder
	ResolveWithAlias(ctx *context.ResolverContext, alias *TableAlias) ([]fields.Field, error)
}

type AliasedTableReferenceOrExpression struct {
	BaseTableExpression TableReferenceOrExpression
	Alias               *TableAlias
}

func (a AliasedTableReferenceOrExpression) String() string {
	return fmt.Sprintf("AliasedTableReferenceOrExpression(%s) AS %s", a.BaseTableExpression, a.Alias)
}

type TableAlias struct {
	TableAlias    string
	ColumnAliases []string
}

func (a *TableAlias) String() string {
	return fmt.Sprintf("TableAlias(%s)(%s)", a.TableAlias, strings.Join(a.ColumnAliases, ", "))
}

type TableReference struct {
	Name  string
	table impls.Table
}

func (r *TableReference) String() string {
	return fmt.Sprintf("TableReference(%s)", r.Name)
}

func (r *TableReference) Resolve(ctx *context.ResolverContext) ([]fields.Field, error) {
	return r.ResolveWithAlias(ctx, nil)
}

func (r *TableReference) ResolveWithAlias(ctx *context.ResolverContext, alias *TableAlias) ([]fields.Field, error) {
	table, ok := ctx.Tables.Get(r.Name)
	if !ok {
		return nil, fmt.Errorf("unknown table %q", r.Name)
	}
	r.table = table

	name := r.Name
	if alias != nil {
		if len(alias.ColumnAliases) != 0 {
			return nil, fmt.Errorf("column aliases unimplemented")
		}

		name = alias.TableAlias
	}

	var exprs []impls.Expression
	var fields []fields.Field
	for _, f := range table.Fields() {
		f2 := f.Field.WithRelationName(name)
		exprs = append(exprs, expressions.NewNamed(f2))
		fields = append(fields, f2)
	}

	if _, err := ctx.SymbolTable.AddRelation(r.Name, name, exprs, fields); err != nil {
		return nil, err
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

func (te TableExpression) String() string {
	return fmt.Sprintf("TableExpression(%s)", te.Base)
}

type Join struct {
	Table     TableExpression
	Condition impls.Expression
}

func (e TableExpression) Resolve(ctx *context.ResolverContext) ([]fields.Field, error) {
	return e.ResolveWithAlias(ctx, nil)
}

func (e TableExpression) ResolveWithAlias(ctx *context.ResolverContext, alias *TableAlias) ([]fields.Field, error) {
	if alias != nil {
		panic("OH NO") // TODO
	}

	fields, err := e.Base.BaseTableExpression.ResolveWithAlias(ctx, e.Base.Alias)
	if err != nil {
		return nil, err
	}

	// if e.Base.Alias != nil {
	// 	if len(e.Base.Alias.ColumnAliases) != 0 {
	// 		return nil, fmt.Errorf("column alias resolution not yet implemented")
	// 	}

	// 	if _, ok := e.Base.BaseTableExpression.(*TableReference); !ok {
	// 		if err := ctx.SymbolTable.AddRelation(e.Base.Alias.TableAlias, fields); err != nil {
	// 			return nil, err
	// 		}
	// 	}
	// }

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
