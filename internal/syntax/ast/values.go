package ast

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/access"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type ValuesBuilder struct {
	Fields      []fields.Field
	Expressions [][]impls.Expression
}

func (b ValuesBuilder) TableExpression() {}

func (b *ValuesBuilder) Resolve(ctx ResolveContext) ([]fields.Field, error) {
	return b.Fields, nil // TODO - check expression types
}

func (b *ValuesBuilder) Build() (queries.Node, error) {
	return access.NewValues(b.Fields, b.Expressions), nil
}
