package nodes

import (
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type deleteNode struct {
	Node
	table     impls.Table
	aliasName string
}

func NewDelete(node Node, table impls.Table, aliasName string) Node {
	return &deleteNode{
		Node:      node,
		table:     table,
		aliasName: aliasName,
	}
}

func (n *deleteNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("delete from %s", n.table.Name())
	n.Node.Serialize(w.Indent())
}

func (n *deleteNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Delete scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Delete")

		row, err := scanner.Scan()
		if err != nil {
			return rows.Row{}, err
		}

		relationName := n.table.Name()
		if n.aliasName != "" {
			relationName = n.aliasName
		}

		tidRow, err := row.IsolateTID(relationName)
		if err != nil {
			return rows.Row{}, err
		}

		deletedRow, ok, err := n.table.Delete(tidRow)
		if err != nil {
			return rows.Row{}, err
		}
		if !ok {
			return rows.Row{}, scan.ErrNoRows
		}

		return deletedRow, nil
	}), nil
}
