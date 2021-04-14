package repl

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/relations"
	"github.com/efritz/gostgres/internal/shared"
)

var builtins = map[string]relations.Relation{
	"A": tableA,
	"B": tableB,
	"C": tableC,
}

var tableA = relations.NewData(
	"A",
	shared.Rows{
		Fields: []shared.Field{{RelationName: "A", Name: "a"}},
		Values: [][]interface{}{
			{6},
			{5},
			{7},
			{3},
			{4},
		},
	},
)

var tableB = relations.NewData(
	"B",
	shared.Rows{
		Fields: []shared.Field{{RelationName: "B", Name: "b"}},
		Values: [][]interface{}{
			{2},
			{3},
			{8},
		},
	},
)

var tableC = relations.NewDataTemp(
	"C",
	shared.Rows{
		Fields: []shared.Field{{RelationName: "C", Name: "c"}},
		Values: [][]interface{}{
			{"89d29124d43623cc32f8d9f999fd"},
			{"33623cc32f8d9f999fd69189d29124d4368c20ab"},
			{"189d29124d4368c20ab33623cc32f8d9fd69"},
			{"0ad2e75d529bda744b07fe7"},
		},
	},
	nil,
	expressions.Int(expressions.NewLength(expressions.NewNamed("", "c"))),
)
