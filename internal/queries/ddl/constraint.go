package ddl

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/constraints"
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/indexes"
	"github.com/efritz/gostgres/internal/protocol"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/shared"
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

func (q *createPrimaryKeyConstraint) Execute(ctx queries.Context, w protocol.ResponseWriter) {
	if err := q.ExecuteDDL(ctx); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}

func (q *createPrimaryKeyConstraint) ExecuteDDL(ctx queries.Context) error {
	table, ok := ctx.Tables.GetTable(q.tableName)
	if !ok {
		return fmt.Errorf("unknown table %q", q.tableName)
	}

	fields := table.Fields()
	var columnExpressions []expressions.ExpressionWithDirection
	for _, columnName := range q.columnNames {
		i := slices.IndexFunc(fields, func(f shared.TableField) bool { return f.Name() == columnName })
		if i < 0 {
			return fmt.Errorf("no such column %q on table %q", columnName, q.tableName)
		}
		field := fields[i]

		if field.Nullable() {
			return fmt.Errorf("primary key fields must be nullable")
		}

		columnExpressions = append(columnExpressions, expressions.ExpressionWithDirection{
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

	return table.SetPrimaryKey(index)
}

type createCheckConstraint struct {
	name       string
	tableName  string
	expression expressions.Expression
}

var _ queries.Query = &createCheckConstraint{}
var _ DDLQuery = &createCheckConstraint{}

func NewCreateCheckConstraint(name, tableName string, expression expressions.Expression) *createCheckConstraint {
	return &createCheckConstraint{
		name:       name,
		tableName:  tableName,
		expression: expression,
	}
}

func (q *createCheckConstraint) Execute(ctx queries.Context, w protocol.ResponseWriter) {
	if err := q.ExecuteDDL(ctx); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}

func (q *createCheckConstraint) ExecuteDDL(ctx queries.Context) error {
	table, ok := ctx.Tables.GetTable(q.tableName)
	if !ok {
		return fmt.Errorf("unknown table %q", q.tableName)
	}

	constraint := constraints.NewCheckConstraint(q.name, q.expression)
	return table.AddConstraint(constraint)
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

func (q *createForeignKeyConstraint) Execute(ctx queries.Context, w protocol.ResponseWriter) {
	if err := q.ExecuteDDL(ctx); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}

func (q *createForeignKeyConstraint) ExecuteDDL(ctx queries.Context) error {
	table, ok := ctx.Tables.GetTable(q.tableName)
	if !ok {
		return fmt.Errorf("unknown table %q", q.tableName)
	}

	var exprs []expressions.Expression
	for _, columnName := range q.columnNames {
		i := slices.IndexFunc(table.Fields(), func(f shared.TableField) bool { return f.Name() == columnName })
		if i < 0 {
			return fmt.Errorf("no such column %q on table %q", columnName, q.tableName)
		}
		field := table.Fields()[i].Field

		exprs = append(exprs, setRelationName(expressions.NewNamed(field), q.tableName))
	}

	refTable, ok := ctx.Tables.GetTable(q.refTableName)
	if !ok {
		return fmt.Errorf("unknown table %q", q.refTableName)
	}

	var refIndex indexes.Index[indexes.BtreeIndexScanOptions]
	for _, index := range refTable.Indexes() {
		btreeIndex, ok := index.(indexes.Index[indexes.BtreeIndexScanOptions])
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
	return table.AddConstraint(constraint)
}
