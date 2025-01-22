package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type AnalyzeBuilder struct {
	TableNames []string
	tables     []impls.Table
}

func (b *AnalyzeBuilder) Resolve(ctx *impls.NodeResolutionContext) error {
	tableNames := b.TableNames
	if len(tableNames) == 0 {
		tableNames = ctx.Catalog().Tables.Keys()
	}

	var tables []impls.Table
	for _, name := range tableNames {
		table, ok := ctx.Catalog().Tables.Get(name)
		if !ok {
			return fmt.Errorf("unknown table %q", name)
		}

		tables = append(tables, table)
	}

	b.tables = tables
	return nil
}

func (b *AnalyzeBuilder) Build() (plan.LogicalNode, error) {
	return plan.NewAnalyze(b.tables), nil
}
