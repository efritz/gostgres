package mutation

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	projectionHelpers "github.com/efritz/gostgres/internal/execution/queries/nodes/projection"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type updateNode struct {
	nodes.Node
	table          impls.Table
	aliasName      string
	setExpressions []SetExpression
	projection     *projection.Projection
}

type SetExpression struct {
	Name       string
	Expression impls.Expression
}

func NewUpdate(
	node nodes.Node,
	table impls.Table,
	aliasName string,
	setExpressions []SetExpression,
	projection *projection.Projection,
) nodes.Node {
	return &updateNode{
		Node:           node,
		table:          table,
		aliasName:      aliasName,
		setExpressions: setExpressions,
		projection:     projection,
	}
}

func (n *updateNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("update %s", n.table.Name())
	n.Node.Serialize(w.Indent())

	if n.projection != nil {
		w.WritefLine("returning %s", n.projection)
		n.Node.Serialize(w.Indent())
	}
}

func (n *updateNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Update scanner")

	updatedRows, err := n.updateRows(ctx)
	if err != nil {
		return nil, err
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Update")

		if len(updatedRows) != 0 {
			return rows.Row{}, scan.ErrNoRows
		}

		row := updatedRows[0]
		updatedRows = updatedRows[1:]
		return projectionHelpers.Project(ctx, row, n.projection)
	}), nil
}

func (n *updateNode) updateRows(ctx impls.ExecutionContext) ([]rows.Row, error) {
	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	var updatedRows []rows.Row
	for {
		row, err := scanner.Scan()
		if err != nil {
			if err == scan.ErrNoRows {
				break
			}

			return nil, err
		}

		updates := map[int]any{}
		for _, set := range n.setExpressions {
			value, err := queries.Evaluate(ctx, set.Expression, row)
			if err != nil {
				return nil, err
			}

			found := false
			for i, field := range n.table.Fields() {
				if field.Name() == set.Name {
					if field.Internal() {
						return nil, fmt.Errorf("cannot update internal field %q", field.Name())
					}

					found = true
					updates[i] = value
					break
				}
			}

			if !found {
				return nil, fmt.Errorf("unknown column %q", set.Name)
			}
		}

		relationName := n.table.Name()
		if n.aliasName != "" {
			relationName = n.aliasName
		}

		tidRow, err := row.IsolateTID(relationName)
		if err != nil {
			return nil, err
		}

		baseRow, ok, err := n.table.Delete(tidRow)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, nil
		}

		for i, value := range updates {
			baseRow.Values[i] = value
		}

		updatedRow, err := n.table.Insert(ctx, baseRow.DropInternalFields())
		if err != nil {
			return nil, err
		}

		if n.projection != nil {
			updatedRows = append(updatedRows, updatedRow)
		}
	}

	return updatedRows, nil
}
