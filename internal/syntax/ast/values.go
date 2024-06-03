package ast

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/access"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type ValuesBuilder struct {
	RowFields         []fields.Field
	AllRowExpressions [][]impls.Expression
}

func (b *ValuesBuilder) Build(ctx BuildContext) (queries.Node, error) {
	return b.TableExpression(ctx)
}

func (b ValuesBuilder) TableExpression(ctx BuildContext) (queries.Node, error) {
	return access.NewValues(b.RowFields, b.AllRowExpressions), nil
}
