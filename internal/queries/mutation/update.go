package mutation

import (
	"fmt"
	"io"
	"slices"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/queries/projection"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/table"
)

type updateNode struct {
	queries.Node
	table          *table.Table
	setExpressions []SetExpression
	columnNames    []string
	projector      *projection.Projector
}

var _ queries.Node = &updateNode{}

type SetExpression struct {
	Name       string
	Expression expressions.Expression
}

func NewUpdate(node queries.Node, table *table.Table, setExpressions []SetExpression, alias string, expressions []projection.ProjectionExpression) (queries.Node, error) {
	if alias != "" {
		for i, pe := range expressions {
			expressions[i] = pe.Dealias(table.Name(), table.Fields(), alias)
		}
	}

	projector, err := projection.NewProjector(node.Name(), table.Fields(), expressions)
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

func (n *updateNode) Fields() []shared.Field {
	return slices.Clone(n.projector.Fields())
}

func (n *updateNode) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%supdate returning (%s)\n", serialization.Indent(indentationLevel), n.projector))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *updateNode) Optimize() {
	n.projector.Optimize()
	n.Node.Optimize()
}

func (n *updateNode) AddFilter(filter expressions.Expression) {
}

func (n *updateNode) AddOrder(order expressions.OrderExpression) {
}

func (n *updateNode) Filter() expressions.Expression {
	return nil
}

func (n *updateNode) Ordering() expressions.OrderExpression {
	return nil
}

func (n *updateNode) SupportsMarkRestore() bool {
	return false
}

func (n *updateNode) Scanner(ctx queries.Context) (scan.Scanner, error) {
	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return scan.ScannerFunc(func() (shared.Row, error) {
		row, err := scanner.Scan()
		if err != nil {
			return shared.Row{}, err
		}

		values := slices.Clone(row.Values)

		for _, set := range n.setExpressions {
			value, err := ctx.Evaluate(set.Expression, row)
			if err != nil {
				return shared.Row{}, err
			}

			found := false
			for i, field := range row.Fields {
				if field.Name() == set.Name {
					if field.Internal() {
						return shared.Row{}, fmt.Errorf("cannot update internal field %s", set.Name)
					}

					found = true
					values[i] = value
				}
			}

			if !found {
				return shared.Row{}, fmt.Errorf("unknown column %s", set.Name)
			}
		}

		deletedRow, err := shared.NewRow(row.Fields[:1], values[:1])
		if err != nil {
			return shared.Row{}, err
		}
		if _, ok, err := n.table.Delete(deletedRow); err != nil {
			return shared.Row{}, err
		} else if !ok {
			return shared.Row{}, nil
		}

		insertedRow, err := shared.NewRow(row.Fields[1:], values[1:])
		if err != nil {
			return shared.Row{}, err
		}

		updatedRow, err := n.table.Insert(insertedRow)
		if err != nil {
			return shared.Row{}, err
		}

		return n.projector.ProjectRow(ctx, updatedRow)
	}), nil
}