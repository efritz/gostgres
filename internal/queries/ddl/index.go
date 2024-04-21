package ddl

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/indexes"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/table"
)

type createIndex struct {
	name              string
	tableName         string
	method            string
	columnExpressions []expressions.ExpressionWithDirection
	where             expressions.Expression
}

var _ queries.Node = &createIndex{}

func NewCreateIndex(name, tableName, method string, columnExpressions []expressions.ExpressionWithDirection, where expressions.Expression) *createIndex {
	return &createIndex{
		name:              name,
		tableName:         tableName,
		method:            method,
		columnExpressions: columnExpressions,
		where:             where,
	}
}

func (n *createIndex) Name() string {
	return "CREATE INDEX"
}

func (n *createIndex) Fields() []shared.Field {
	return nil
}

func (n *createIndex) Serialize(w io.Writer, indentationLevel int) {
	// TODO - implement for DDL?
}

func (n *createIndex) Optimize() {
}

func (n *createIndex) AddFilter(filter expressions.Expression) {
}

func (n *createIndex) AddOrder(order expressions.OrderExpression) {
}

func (n *createIndex) Filter() expressions.Expression {
	return nil
}

func (n *createIndex) Ordering() expressions.OrderExpression {
	return nil
}

func (n *createIndex) SupportsMarkRestore() bool {
	return false
}

// TODO - HACK
type temp interface {
	GetTable(name string) (*table.Table, bool)
}

func (n *createIndex) Scanner(ctx scan.ScanContext) (scan.Scanner, error) {
	t, ok := ctx.Tables.(temp)
	if !ok {
		panic("OH NO")
	}

	table, ok := t.GetTable(n.tableName)
	if !ok {
		return nil, fmt.Errorf("unknown table %q", n.tableName)
	}

	factories := map[string]func() (indexes.BaseIndex, error){
		"btree": n.createBtreeIndex,
		"hash":  n.createHashIndex,
	}

	factory, ok := factories[n.method]
	if !ok {
		return nil, fmt.Errorf("unknown index method %q", n.method)
	}

	index, err := factory()
	if err != nil {
		return nil, err
	}

	if err := table.AddIndex(index); err != nil {
		return nil, err
	}

	return scan.ScannerFunc(func() (shared.Row, error) {
		return shared.Row{}, scan.ErrNoRows
	}), nil
}

func (n *createIndex) createBtreeIndex() (indexes.BaseIndex, error) {
	var index indexes.Index[indexes.BtreeIndexScanOptions] = indexes.NewBTreeIndex(
		n.name,
		n.tableName,
		n.columnExpressions,
	)
	if n.where != nil {
		index = indexes.NewPartialIndex(index, n.where)
	}

	return index, nil
}

func (n *createIndex) createHashIndex() (indexes.BaseIndex, error) {
	if len(n.columnExpressions) != 1 {
		return nil, fmt.Errorf("hash index must have exactly one column")
	}
	if n.columnExpressions[0].Reverse {
		return nil, fmt.Errorf("hash index do not support ordering")
	}

	var index indexes.Index[indexes.HashIndexScanOptions] = indexes.NewHashIndex(
		n.name,
		n.tableName,
		n.columnExpressions[0].Expression,
	)
	if n.where != nil {
		index = indexes.NewPartialIndex(index, n.where)
	}

	return index, nil
}
