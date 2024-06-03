package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/mutation"
	"github.com/efritz/gostgres/internal/execution/queries/projection"
)

type InsertBuilder struct {
	TableDescription     TableDescription
	ColumnNames          []string
	Node                 BaseTableExpressionDescription
	ReturningExpressions []projection.ProjectionExpression
}

func (b *InsertBuilder) Build(ctx BuildContext) (queries.Node, error) {
	node, err := b.Node.Build(ctx)
	if err != nil {
		return nil, err
	}

	table, ok := ctx.Tables.Get(b.TableDescription.Name)
	if !ok {
		return nil, fmt.Errorf("unknown table %s", b.TableDescription.Name)
	}

	return mutation.NewInsert(node, table, b.TableDescription.Name, b.TableDescription.AliasName, b.ColumnNames, b.ReturningExpressions)
}
