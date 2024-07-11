package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/projector"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/mutation"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/ast/context"
)

type InsertBuilder struct {
	Target      TargetTable
	ColumnNames []string
	Source      TableReferenceOrExpression
	Returning   []projector.ProjectionExpression
	table       impls.Table
}

func (b *InsertBuilder) Resolve(ctx *context.ResolverContext) ([]fields.Field, error) {
	table, ok := ctx.Tables.Get(b.Target.Name)
	if !ok {
		return nil, fmt.Errorf("unknown table %q", b.Target.Name)
	}
	b.table = table

	ctx.SymbolTable.PushScope()
	defer ctx.SymbolTable.PopScope()

	fs, err := b.Source.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	// name := b.Target.AliasName
	// if name == "" {
	// 	name = b.Target.Name
	// }
	// if _, err := ctx.SymbolTable.AddRelation(name, fs); err != nil {
	// 	return nil, err
	// }

	// TODO - returning n stuff
	return fs, nil
}

func (b *InsertBuilder) Build() (queries.Node, error) {
	node, err := b.Source.Build()
	if err != nil {
		return nil, err
	}

	return mutation.NewInsert(node, b.table, b.Target.Name, b.Target.AliasName, b.ColumnNames, b.Returning)
}
