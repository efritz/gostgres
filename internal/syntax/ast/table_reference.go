package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

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

func (r *TableReference) Build() (plan.LogicalNode, error) {
	return plan.NewAccess(r.table), nil
}
