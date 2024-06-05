package mutation

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/projector"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type updateNode struct {
	queries.Node
	table          impls.Table
	setExpressions []SetExpression
	columnNames    []string
	projector      *projector.Projector
}

var _ queries.Node = &updateNode{}

type SetExpression struct {
	Name       string
	Expression impls.Expression
}

func NewUpdate(node queries.Node, table impls.Table, setExpressions []SetExpression, alias string, expressions []projector.ProjectionExpression) (queries.Node, error) {
	var fields []fields.Field
	for _, field := range table.Fields() {
		fields = append(fields, field.Field)
	}

	if alias != "" {
		for i, pe := range expressions {
			expressions[i] = pe.Dealias(table.Name(), fields, alias)
		}
	}

	projector, err := projector.NewProjector(node.Name(), fields, expressions)
	if err != nil {
		return nil, err
	}

	return &updateNode{
		Node:           node,
		table:          table,
		setExpressions: setExpressions,
		projector:      projector,
	}, nil
}

func (n *updateNode) Fields() []fields.Field {
	return slices.Clone(n.projector.Fields())
}

func (n *updateNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("update returning (%s)", n.projector)
	n.Node.Serialize(w.Indent())
}

func (n *updateNode) AddFilter(filter impls.Expression)    {}
func (n *updateNode) AddOrder(order impls.OrderExpression) {}

func (n *updateNode) Optimize() {
	n.projector.Optimize()
	n.Node.Optimize()
}

func (n *updateNode) Filter() impls.Expression        { return nil }
func (n *updateNode) Ordering() impls.OrderExpression { return nil }
func (n *updateNode) SupportsMarkRestore() bool       { return false }

func (n *updateNode) Scanner(ctx impls.Context) (scan.RowScanner, error) {
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

		values := slices.Clone(row.Values)

		for _, set := range n.setExpressions {
			value, err := queries.Evaluate(ctx, set.Expression, row)
			if err != nil {
				return rows.Row{}, err
			}

			found := false
			for i, field := range row.Fields {
				if field.Name() == set.Name {
					if field.Internal() {
						return rows.Row{}, fmt.Errorf("cannot update internal field %s", set.Name)
					}

					found = true
					values[i] = value
				}
			}

			if !found {
				return rows.Row{}, fmt.Errorf("unknown column %s", set.Name)
			}
		}

		deletedRow, err := rows.NewRow(row.Fields[:1], values[:1])
		if err != nil {
			return rows.Row{}, err
		}
		if _, ok, err := n.table.Delete(deletedRow); err != nil {
			return rows.Row{}, err
		} else if !ok {
			return rows.Row{}, nil
		}

		insertedRow, err := rows.NewRow(row.Fields[1:], values[1:])
		if err != nil {
			return rows.Row{}, err
		}

		updatedRow, err := n.table.Insert(ctx, insertedRow)
		if err != nil {
			return rows.Row{}, err
		}

		return n.projector.ProjectRow(ctx, updatedRow)
	}), nil
}
