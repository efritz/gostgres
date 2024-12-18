package impls

import "github.com/efritz/gostgres/internal/shared/rows"

type Table interface {
	Name() string
	Indexes() []BaseIndex
	Fields() []TableField
	TIDs() []int64
	Row(tid int64) (rows.Row, bool)
	SetPrimaryKey(index BaseIndex) error
	AddIndex(index BaseIndex) error
	AddConstraint(ctx ExecutionContext, constraint Constraint) error
	Insert(ctx ExecutionContext, row rows.Row) (_ rows.Row, err error)
	Delete(row rows.Row) (rows.Row, bool, error)
	Statistics() (TableStatistics, []ColumnStatistics)
	Analyze() error
}
