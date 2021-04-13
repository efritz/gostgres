package repl

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/relations"
	"github.com/efritz/gostgres/internal/shared"
)

var tableA = relations.NewData(
	"A",
	shared.Rows{
		Fields: []shared.Field{{Name: "a"}},
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
		Fields: []shared.Field{{Name: "b"}},
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
		Fields: []shared.Field{{Name: "c"}},
		Values: [][]interface{}{
			{"89d29124d43623cc32f8d9f999fd"},
			{"33623cc32f8d9f999fd69189d29124d4368c20ab"},
			{"189d29124d4368c20ab33623cc32f8d9fd69"},
			{"0ad2e75d529bda744b07fe7"},
		},
	},
	nil, // filters.NewEquals(expressions.NewNamed("", "c"), expressions.NewConstant("33623cc32f8d9f999fd69189d29124d4368c20ab")),
	expressions.Int(expressions.NewLength(expressions.NewNamed("", "c"))),
)

var builtins = map[string]relations.Relation{
	"A": tableA,
	"B": tableB,
	"C": tableC,
}

func parseRelation(text string) (relations.Relation, error) {
	return relations.NewLimit(
		// relations.NewOffset(
		// relations.NewOrder(
		relations.NewFilter(
			relations.NewAlias(
				relations.NewProjection(
					relations.NewJoin(
						relations.NewJoin(
							tableA,
							tableB,
							nil,
						),
						tableC,
						nil,
					),
					[]relations.AliasedExpression{
						{Alias: "id", Expression: expressions.NewNamed("A", "a")},
						{Alias: "repository_id", Expression: expressions.NewNamed("B", "b")},
						{Alias: "commit", Expression: expressions.NewNamed("C", "c")},
					},
				),
				"Q",
			),
			expressions.Bool(expressions.NewEquals(
				expressions.NewNamed("", "id"),
				expressions.NewSum(
					expressions.NewNamed("", "repository_id"),
					expressions.NewNamed("", "repository_id"),
				),
			),
			)),
		// 	expressions.NewNamed("", "id"),
		// ),
		// 	1,
		// ),
		5,
	), nil
}

func Start() error {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		command, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}
		if command == "exit" {
			return nil
		}

		relation, err := parseRelation(command)
		if err != nil {
			fmt.Printf("failed to parse relation: %s", err)
			continue
		}

		rows, err := relations.ScanRows(relation)
		if err != nil {
			fmt.Printf("failed to execute query: %s", err)
			continue
		}

		displayValues(rows)
	}
}
