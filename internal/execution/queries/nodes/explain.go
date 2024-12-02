package nodes

import (
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type explainNode struct {
	n      Node
	fields []fields.Field
}

func NewExplain(n Node, fields []fields.Field) Node {
	return &explainNode{
		n:      n,
		fields: fields,
	}
}

func (n *explainNode) Serialize(w serialization.IndentWriter) {
	// Explain never explained
}

func (n *explainNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Explain scanner")

	plan := serialization.SerializePlan(n.n)
	emitted := false

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Explain")

		if emitted {
			return rows.Row{}, scan.ErrNoRows
		}

		emitted = true
		return rows.NewRow(n.fields, []any{plan})
	}), nil
}
