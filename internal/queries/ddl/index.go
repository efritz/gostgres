package ddl

import (
	"fmt"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/indexes"
	"github.com/efritz/gostgres/internal/protocol"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/table"
)

type createIndex struct {
	name              string
	tableName         string
	method            string
	columnExpressions []expressions.ExpressionWithDirection
	where             expressions.Expression
}

var _ queries.Query = &createIndex{}

func NewCreateIndex(name, tableName, method string, columnExpressions []expressions.ExpressionWithDirection, where expressions.Expression) *createIndex {
	return &createIndex{
		name:              name,
		tableName:         tableName,
		method:            method,
		columnExpressions: columnExpressions,
		where:             where,
	}
}

func (n *createIndex) Execute(ctx queries.Context, w protocol.ResponseWriter) {
	if err := n.execute(ctx); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}

func (n *createIndex) execute(ctx queries.Context) error {
	factories := map[string]func() (table.Index, error){
		"btree": n.createBtreeIndex,
		"hash":  n.createHashIndex,
	}

	factory, ok := factories[n.method]
	if !ok {
		return fmt.Errorf("unknown index method %q", n.method)
	}

	index, err := factory()
	if err != nil {
		return err
	}

	table, ok := ctx.Tables.GetTable(n.tableName)
	if !ok {
		return fmt.Errorf("unknown table %q", n.tableName)
	}

	if err := table.AddIndex(index); err != nil {
		return err
	}

	return nil
}

func (n *createIndex) createBtreeIndex() (table.Index, error) {
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

func (n *createIndex) createHashIndex() (table.Index, error) {
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
