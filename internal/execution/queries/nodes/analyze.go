package nodes

import (
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type analyzeNode struct {
	tables []impls.Table
}

func NewAnalyze(tables []impls.Table) Node {
	return &analyzeNode{
		tables: tables,
	}
}

func (n *analyzeNode) Serialize(w serialization.IndentWriter) {
	// Analyze never explained
}

func (n *analyzeNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Analyze scanner")

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Analyze")

		for _, table := range n.tables {
			ctx.Log("Analyzing table %s", table.Name())

			if err := table.Analyze(); err != nil {
				return rows.Row{}, err
			}
		}

		return rows.Row{}, scan.ErrNoRows
	}), nil
}
