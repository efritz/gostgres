package ddl

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/catalog/table/constraints"
	"github.com/efritz/gostgres/internal/catalog/table/indexes"
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type createPrimaryKeyConstraint struct {
	name        string
	tableName   string
	columnNames []string
}

var _ queries.Query = &createPrimaryKeyConstraint{}
var _ DDLQuery = &createPrimaryKeyConstraint{}

func NewCreatePrimaryKeyConstraint(name, tableName string, columnNames []string) *createPrimaryKeyConstraint {
	return &createPrimaryKeyConstraint{
		name:        name,
		tableName:   tableName,
		columnNames: columnNames,
	}
}

func (q *createPrimaryKeyConstraint) Execute(ctx impls.ExecutionContext, w protocol.ResponseWriter) {
	if err := q.ExecuteDDL(ctx); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}

func (q *createPrimaryKeyConstraint) ExecuteDDL(ctx impls.ExecutionContext) error {
	t, ok := ctx.Catalog().Tables.Get(q.tableName)
	if !ok {
		return fmt.Errorf("unknown table %q", q.tableName)
	}

	fields := t.Fields()
	var columnExpressions []impls.ExpressionWithDirection
	for _, columnName := range q.columnNames {
		i := slices.IndexFunc(fields, func(f impls.TableField) bool { return f.Name() == columnName })
		if i < 0 {
			return fmt.Errorf("no such column %q on table %q", columnName, q.tableName)
		}
		field := fields[i]

		if field.Nullable() {
			return fmt.Errorf("primary key fields must be nullable")
		}

		columnExpressions = append(columnExpressions, impls.ExpressionWithDirection{
			Expression: setRelationName(expressions.NewNamed(field.Field), q.tableName),
			Reverse:    false,
		})
	}

	index := indexes.NewBTreeIndex(
		q.name,
		q.tableName,
		true,
		columnExpressions,
	)

	return t.SetPrimaryKey(index)
}

type createCheckConstraint struct {
	name       string
	tableName  string
	expression impls.Expression
}

var _ queries.Query = &createCheckConstraint{}
var _ DDLQuery = &createCheckConstraint{}

func NewCreateCheckConstraint(name, tableName string, expression impls.Expression) *createCheckConstraint {
	return &createCheckConstraint{
		name:       name,
		tableName:  tableName,
		expression: expression,
	}
}

func (q *createCheckConstraint) Execute(ctx impls.ExecutionContext, w protocol.ResponseWriter) {
	if err := q.ExecuteDDL(ctx); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}

func (q *createCheckConstraint) ExecuteDDL(ctx impls.ExecutionContext) error {
	table, ok := ctx.Catalog().Tables.Get(q.tableName)
	if !ok {
		return fmt.Errorf("unknown table %q", q.tableName)
	}

	constraint := constraints.NewCheckConstraint(q.name, q.expression)
	return table.AddConstraint(ctx, constraint)
}

type createForeignKeyConstraint struct {
	name           string
	tableName      string
	columnNames    []string
	refTableName   string
	refColumnNames []string
}

var _ queries.Query = &createForeignKeyConstraint{}
var _ DDLQuery = &createForeignKeyConstraint{}

func NewCreateForeignKeyConstraint(name, tableName string, columnNames []string, refTableName string, refColumnNames []string) *createForeignKeyConstraint {
	return &createForeignKeyConstraint{
		name:           name,
		tableName:      tableName,
		columnNames:    columnNames,
		refTableName:   refTableName,
		refColumnNames: refColumnNames,
	}
}

func (q *createForeignKeyConstraint) Execute(ctx impls.ExecutionContext, w protocol.ResponseWriter) {
	if err := q.ExecuteDDL(ctx); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}

func (q *createForeignKeyConstraint) ExecuteDDL(ctx impls.ExecutionContext) error {
	t, ok := ctx.Catalog().Tables.Get(q.tableName)
	if !ok {
		return fmt.Errorf("unknown table %q", q.tableName)
	}

	var exprs []impls.Expression
	for _, columnName := range q.columnNames {
		i := slices.IndexFunc(t.Fields(), func(f impls.TableField) bool { return f.Name() == columnName })
		if i < 0 {
			return fmt.Errorf("no such column %q on table %q", columnName, q.tableName)
		}
		field := t.Fields()[i].Field

		exprs = append(exprs, setRelationName(expressions.NewNamed(field), q.tableName))
	}

	refTable, ok := ctx.Catalog().Tables.Get(q.refTableName)
	if !ok {
		return fmt.Errorf("unknown table %q", q.refTableName)
	}

	var refIndex impls.Index[indexes.BtreeIndexScanOptions]
	for _, index := range refTable.Indexes() {
		btreeIndex, ok := index.(impls.Index[indexes.BtreeIndexScanOptions])
		if !ok {
			continue
		}

		var fieldNames []string
		for _, field := range index.UniqueOn() {
			fieldNames = append(fieldNames, field.Name())
		}

		if len(fieldNames) > 0 && slices.Equal(fieldNames, q.refColumnNames) {
			refIndex = btreeIndex
			break
		}
	}
	if refIndex == nil {
		return fmt.Errorf("there is no unique constraint matching given keys for referenced table")
	}

	constraint := constraints.NewForeignKeyConstraint(q.name, exprs, refIndex)
	return t.AddConstraint(ctx, constraint)
}
