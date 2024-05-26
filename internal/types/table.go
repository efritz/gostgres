package types

import "github.com/efritz/gostgres/internal/shared"

type Table interface {
	Name() string
	Indexes() []BaseIndex
	Fields() []TableField
	Size() int
	TIDs() []int64
	Row(tid int64) (shared.Row, bool)
	SetPrimaryKey(index BaseIndex) error
	AddIndex(index BaseIndex) error
	AddConstraint(ctx Context, constraint Constraint) error
	Insert(ctx Context, row shared.Row) (_ shared.Row, err error)
	Delete(row shared.Row) (shared.Row, bool, error)
}
