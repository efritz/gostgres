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
	unique            bool
	columnExpressions []expressions.ExpressionWithDirection
	where             expressions.Expression
}

var _ queries.Query = &createIndex{}
var _ DDLQuery = &createIndex{}

func NewCreateIndex(name, tableName, method string, unique bool, columnExpressions []expressions.ExpressionWithDirection, where expressions.Expression) *createIndex {
	return &createIndex{
		name:              name,
		tableName:         tableName,
		method:            method,
		unique:            unique,
		columnExpressions: columnExpressions,
		where:             where,
	}
}

func (q *createIndex) Execute(ctx queries.Context, w protocol.ResponseWriter) {
	if err := q.ExecuteDDL(ctx); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}

func (q *createIndex) ExecuteDDL(ctx queries.Context) error {
	factories := map[string]func(ctx queries.Context) (table.Index, error){
		"btree": q.createBtreeIndex,
		"hash":  q.createHashIndex,
	}

	factory, ok := factories[q.method]
	if !ok {
		return fmt.Errorf("unknown index method %q", q.method)
	}

	index, err := factory(ctx)
	if err != nil {
		return err
	}

	table, ok := ctx.Tables.GetTable(q.tableName)
	if !ok {
		return fmt.Errorf("unknown table %q", q.tableName)
	}

	if err := table.AddIndex(index); err != nil {
		return err
	}

	return nil
}

func (q *createIndex) createBtreeIndex(ctx queries.Context) (table.Index, error) {
	var columnExpressions []expressions.ExpressionWithDirection
	for _, column := range q.columnExpressions {
		columnExpressions = append(columnExpressions, expressions.ExpressionWithDirection{
			Expression: setRelationName(column.Expression, q.tableName),
			Reverse:    column.Reverse,
		})
	}

	var index indexes.Index[indexes.BtreeIndexScanOptions] = indexes.NewBTreeIndex(
		q.name,
		q.tableName,
		q.unique,
		columnExpressions,
	)
	if q.where != nil {
		index = indexes.NewPartialIndex(index, setRelationName(q.where, q.tableName))
	}

	return index, nil
}

func (q *createIndex) createHashIndex(ctx queries.Context) (table.Index, error) {
	if len(q.columnExpressions) != 1 {
		return nil, fmt.Errorf("hash index must have exactly one column")
	}
	if q.columnExpressions[0].Reverse {
		return nil, fmt.Errorf("hash index do not support ordering")
	}
	if q.unique {
		return nil, fmt.Errorf("hash index do not support uniqueness")
	}

	var index indexes.Index[indexes.HashIndexScanOptions] = indexes.NewHashIndex(
		q.name,
		q.tableName,
		setRelationName(q.columnExpressions[0].Expression, q.tableName),
	)

	if q.where != nil {
		index = indexes.NewPartialIndex(index, setRelationName(q.where, q.tableName))
	}

	return index, nil
}

func setRelationName(e expressions.Expression, name string) expressions.Expression {
	return e.Map(func(e expressions.Expression) expressions.Expression {
		if f, ok := e.Named(); ok {
			return expressions.NewNamed(f.WithRelationName(name))
		}

		return e
	})
}
