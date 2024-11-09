package ddl

import (
	"fmt"

	"github.com/efritz/gostgres/internal/catalog/table/indexes"
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type createIndex struct {
	name              string
	tableName         string
	method            string
	unique            bool
	columnExpressions []impls.ExpressionWithDirection
	where             impls.Expression
}

var _ queries.Query = &createIndex{}
var _ DDLQuery = &createIndex{}

func NewCreateIndex(name, tableName, method string, unique bool, columnExpressions []impls.ExpressionWithDirection, where impls.Expression) *createIndex {
	return &createIndex{
		name:              name,
		tableName:         tableName,
		method:            method,
		unique:            unique,
		columnExpressions: columnExpressions,
		where:             where,
	}
}

func (q *createIndex) Execute(ctx impls.ExecutionContext, w protocol.ResponseWriter) {
	if err := q.ExecuteDDL(ctx); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}

func (q *createIndex) ExecuteDDL(ctx impls.ExecutionContext) error {
	factories := map[string]func(ctx impls.ExecutionContext) (impls.BaseIndex, error){
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

	table, ok := ctx.Catalog.Tables.Get(q.tableName)
	if !ok {
		return fmt.Errorf("unknown table %q", q.tableName)
	}

	if err := table.AddIndex(index); err != nil {
		return err
	}

	return nil
}

func (q *createIndex) createBtreeIndex(ctx impls.ExecutionContext) (impls.BaseIndex, error) {
	var columnExpressions []impls.ExpressionWithDirection
	for _, column := range q.columnExpressions {
		columnExpressions = append(columnExpressions, impls.ExpressionWithDirection{
			Expression: setRelationName(column.Expression, q.tableName),
			Reverse:    column.Reverse,
		})
	}

	var index impls.Index[indexes.BtreeIndexScanOptions] = indexes.NewBTreeIndex(
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

func (q *createIndex) createHashIndex(ctx impls.ExecutionContext) (impls.BaseIndex, error) {
	if len(q.columnExpressions) != 1 {
		return nil, fmt.Errorf("hash index must have exactly one column")
	}
	if q.columnExpressions[0].Reverse {
		return nil, fmt.Errorf("hash index do not support ordering")
	}
	if q.unique {
		return nil, fmt.Errorf("hash index do not support uniqueness")
	}

	var index impls.Index[indexes.HashIndexScanOptions] = indexes.NewHashIndex(
		q.name,
		q.tableName,
		setRelationName(q.columnExpressions[0].Expression, q.tableName),
	)

	if q.where != nil {
		index = indexes.NewPartialIndex(index, setRelationName(q.where, q.tableName))
	}

	return index, nil
}

func setRelationName(e impls.Expression, name string) impls.Expression {
	mapped, _ := e.Map(func(e impls.Expression) (impls.Expression, error) {
		if named, ok := e.(expressions.NamedExpression); ok {
			return expressions.NewNamed(named.Field().WithRelationName(name)), nil
		}

		return e, nil
	})

	return mapped
}
