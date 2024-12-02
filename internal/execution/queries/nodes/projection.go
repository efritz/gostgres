package nodes

import (
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type projectionNode struct {
	Node
	projection *projection.Projection
}

func NewProjection(node Node, projection *projection.Projection) Node {
	return &projectionNode{
		Node:       node,
		projection: projection,
	}
}

func (n *projectionNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("project %s", n.projection)
	n.Node.Serialize(w.Indent())
}

func (n *projectionNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Projection scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	aliases := n.projection.Aliases()

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Projection")

		row, err := scanner.Scan()
		if err != nil {
			return rows.Row{}, err
		}

		values := make([]any, 0, len(aliases))
		for _, field := range aliases {
			value, err := queries.Evaluate(ctx, field.Expression, row)
			if err != nil {
				return rows.Row{}, err
			}

			values = append(values, value)
		}

		return rows.NewRow(n.projection.Fields(), values)
	}), nil
}
