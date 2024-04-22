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

type deleteNode struct {
	queries.Node
	table       *table.Table
	columnNames []string
	projector   *projection.Projector
}

var _ queries.Node = &deleteNode{}

func NewDelete(node queries.Node, table *table.Table, alias string, expressions []projection.ProjectionExpression) (queries.Node, error) {
	if alias != "" {
		for i, pe := range expressions {
			expressions[i] = pe.Dealias(table.Name(), table.Fields(), alias)
		}
	}

	projector, err := projection.NewProjector(node.Name(), table.Fields(), expressions)
	if err != nil {
		return nil, err
	}

	return &deleteNode{
		Node:      node,
		table:     table,
		projector: projector,
	}, nil
}

func (n *deleteNode) Fields() []shared.Field {
	return slices.Clone(n.projector.Fields())
}

func (n *deleteNode) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%sdelete returning (%s)\n", serialization.Indent(indentationLevel), n.projector))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *deleteNode) Optimize() {
	n.projector.Optimize()
	n.Node.Optimize()
}

func (n *deleteNode) AddFilter(filter expressions.Expression) {
}

func (n *deleteNode) AddOrder(order expressions.OrderExpression) {
}

func (n *deleteNode) Filter() expressions.Expression {
	return nil
}

func (n *deleteNode) Ordering() expressions.OrderExpression {
	return nil
}

func (n *deleteNode) SupportsMarkRestore() bool {
	return false
}

func (n *deleteNode) Scanner(ctx queries.Context) (scan.Scanner, error) {
	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return scan.ScannerFunc(func() (shared.Row, error) {
		row, err := scanner.Scan()
		if err != nil {
			return shared.Row{}, err
		}

		deletedRow, ok, err := n.table.Delete(row)
		if err != nil {
			return shared.Row{}, err
		}
		if !ok {
			return shared.Row{}, scan.ErrNoRows
		}

		return n.projector.ProjectRow(ctx, deletedRow)
	}), nil
}
