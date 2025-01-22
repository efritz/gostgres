package mutation

import (
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	projectionHelpers "github.com/efritz/gostgres/internal/execution/queries/nodes/projection"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type deleteNode struct {
	nodes.Node
	table      impls.Table
	aliasName  string
	projection *projection.Projection
}

func NewDelete(node nodes.Node, table impls.Table, aliasName string, projection *projection.Projection) nodes.Node {
	return &deleteNode{
		Node:       node,
		table:      table,
		aliasName:  aliasName,
		projection: projection,
	}
}

func (n *deleteNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("delete from %s", n.table.Name())
	n.Node.Serialize(w.Indent())

	if n.projection != nil {
		w.WritefLine("returning %s", n.projection)
		n.Node.Serialize(w.Indent())
	}
}

func (n *deleteNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Delete scanner")

	deletedRows, err := n.deleteRows(ctx)
	if err != nil {
		return nil, err
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Delete")

		if len(deletedRows) != 0 {
			return rows.Row{}, scan.ErrNoRows
		}

		row := deletedRows[0]
		deletedRows = deletedRows[1:]
		return projectionHelpers.Project(ctx, row, n.projection)
	}), nil
}

func (n *deleteNode) deleteRows(ctx impls.ExecutionContext) ([]rows.Row, error) {
	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	var deletedRows []rows.Row
	for {
		row, err := scanner.Scan()
		if err != nil {
			if err == scan.ErrNoRows {
				break
			}

			return nil, err
		}

		relationName := n.table.Name()
		if n.aliasName != "" {
			relationName = n.aliasName
		}

		tidRow, err := row.IsolateTID(relationName)
		if err != nil {
			return nil, err
		}

		deletedRow, ok, err := n.table.Delete(tidRow)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		if n.projection != nil {
			deletedRows = append(deletedRows, deletedRow)
		}
	}

	return deletedRows, nil
}
