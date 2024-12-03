package nodes

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type updateNode struct {
	Node
	table          impls.Table
	aliasName      string
	setExpressions []SetExpression
}

type SetExpression struct {
	Name       string
	Expression impls.Expression
}

func NewUpdate(
	node Node,
	table impls.Table,
	aliasName string,
	setExpressions []SetExpression,
) Node {
	return &updateNode{
		Node:           node,
		table:          table,
		aliasName:      aliasName,
		setExpressions: setExpressions,
	}
}

func (n *updateNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("update %s", n.table.Name())
	n.Node.Serialize(w.Indent())
}

func (n *updateNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Update scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Update")

		row, err := scanner.Scan()
		if err != nil {
			return rows.Row{}, err
		}

		updates := map[int]any{}
		for _, set := range n.setExpressions {
			value, err := queries.Evaluate(ctx, set.Expression, row)
			if err != nil {
				return rows.Row{}, err
			}

			found := false
			for i, field := range n.table.Fields() {
				if field.Name() == set.Name {
					if field.Internal() {
						return rows.Row{}, fmt.Errorf("cannot update internal field %q", field.Name())
					}

					found = true
					updates[i] = value
					break
				}
			}

			if !found {
				return rows.Row{}, fmt.Errorf("unknown column %q", set.Name)
			}
		}

		relationName := n.table.Name()
		if n.aliasName != "" {
			relationName = n.aliasName
		}

		tidRow, err := row.IsolateTID(relationName)
		if err != nil {
			return rows.Row{}, err
		}

		baseRow, ok, err := n.table.Delete(tidRow)
		if err != nil {
			return rows.Row{}, err
		} else if !ok {
			return rows.Row{}, nil
		}

		for i, value := range updates {
			baseRow.Values[i] = value
		}

		updatedRow, err := n.table.Insert(ctx, baseRow.DropInternalFields())
		if err != nil {
			return rows.Row{}, err
		}

		return updatedRow, nil
	}), nil
}
