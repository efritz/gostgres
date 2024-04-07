package nodes

import "github.com/efritz/gostgres/internal/shared"

type HashIndex struct {
	name    string
	table   *Table
	fields  []shared.Field
	entries map[string]int
}

var _ Index = &HashIndex{}

func NewHashIndex(name string, table *Table, fields []shared.Field) *HashIndex {
	return &HashIndex{
		name:    name,
		table:   table,
		fields:  fields,
		entries: map[string]int{},
	}
}

// TODO
