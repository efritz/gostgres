package ddl

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
)

type createTable struct {
	name   string
	fields []shared.Field
}

var _ queries.Node = &createTable{}

func NewCreateTable(name string, fields []shared.Field) *createTable {
	return &createTable{
		name:   name,
		fields: fields,
	}
}

func (n *createTable) Name() string {
	return "CREATE TABLE"
}

func (n *createTable) Fields() []shared.Field {
	return nil
}

func (n *createTable) Serialize(w io.Writer, indentationLevel int) {
	// TODO - implement for DDL?
}

func (n *createTable) Optimize() {
}

func (n *createTable) AddFilter(filter expressions.Expression) {
}

func (n *createTable) AddOrder(order expressions.OrderExpression) {
}

func (n *createTable) Filter() expressions.Expression {
	return nil
}

func (n *createTable) Ordering() expressions.OrderExpression {
	return nil
}

func (n *createTable) SupportsMarkRestore() bool {
	return false
}

func (n *createTable) Scanner(ctx scan.ScanContext) (scan.Scanner, error) {
	// TODO - create table

	fmt.Printf("> %s: %#v\n", n.name, n.fields)

	return scan.ScannerFunc(func() (shared.Row, error) {
		return shared.Row{}, scan.ErrNoRows
	}), nil
}
