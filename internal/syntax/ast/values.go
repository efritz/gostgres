package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/access"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type ValuesBuilder struct {
	Fields      []fields.Field
	Expressions [][]impls.Expression
}

func (b *ValuesBuilder) Resolve(ctx ResolveContext) error {
	return fmt.Errorf("values resolve unimplemented")
}

func (b ValuesBuilder) TableExpression() {}

func (b *ValuesBuilder) Build(ctx BuildContext) (queries.Node, error) {
	return access.NewValues(b.Fields, b.Expressions), nil
}
