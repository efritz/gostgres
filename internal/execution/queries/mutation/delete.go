package mutation

import (
	"slices"

	"github.com/efritz/gostgres/internal/catalog/table"
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
)

type deleteNode struct {
	queries.Node
	table       *table.Table
	columnNames []string
	projector   *projection.Projector
}

var _ queries.Node = &deleteNode{}

func NewDelete(node queries.Node, table *table.Table, alias string, expressions []projection.ProjectionExpression) (queries.Node, error) {
	var fields []shared.Field
	for _, field := range table.Fields() {
		fields = append(fields, field.Field)
	}

	if alias != "" {
		for i, pe := range expressions {
			expressions[i] = pe.Dealias(table.Name(), fields, alias)
		}
	}

	projector, err := projection.NewProjector(node.Name(), fields, expressions)
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

func (n *deleteNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("delete returning (%s)", n.projector)
	n.Node.Serialize(w.Indent())
}

func (n *deleteNode) AddFilter(filter expressions.Expression)    {}
func (n *deleteNode) AddOrder(order expressions.OrderExpression) {}

func (n *deleteNode) Optimize() {
	n.projector.Optimize()
	n.Node.Optimize()
}

func (n *deleteNode) Filter() expressions.Expression        { return nil }
func (n *deleteNode) Ordering() expressions.OrderExpression { return nil }
func (n *deleteNode) SupportsMarkRestore() bool             { return false }

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
