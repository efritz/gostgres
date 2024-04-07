package nodes

import "github.com/efritz/gostgres/internal/shared"

type BTreeIndex struct {
	name    string
	table   *Table
	fields  []shared.Field
	entries []btreeNode
}

// TODO - make an actual node
type btreeNode struct {
	values []interface{}
	tid    int
}

var _ Index = &BTreeIndex{}

func NewBTreeIndex(name string, table *Table, fields []shared.Field) *BTreeIndex {
	return &BTreeIndex{
		name:    name,
		table:   table,
		fields:  fields,
		entries: nil,
	}
}

// TODO
