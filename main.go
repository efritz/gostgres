package main

import (
	"fmt"
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/filters"
	"github.com/efritz/gostgres/internal/relations"
	"github.com/efritz/gostgres/internal/shared"
)

func main() {
	tableA := relations.NewData(
		"A",
		[]string{"a"},
		[][]interface{}{
			{6},
			{5},
			{7},
			{3},
			{4},
		},
	)
	tableB := relations.NewData(
		"B",
		[]string{"b"},
		[][]interface{}{
			{2},
			{3},
			{8},
		},
	)
	tableC := relations.NewData(
		"C",
		[]string{"c"},
		[][]interface{}{
			{"89d29124d4368c20ab33623cc32f8d9f999fd691"},
			{"189d29124d4368c20ab33623cc32f8d9f999fd69"},
			{"0ad2e75d529bda74472a1dbb5e488ec095b07fe7"},
			{"33623cc32f8d9f999fd69189d29124d4368c20ab"},
		},
	)

	relation := relations.NewLimit(
		relations.NewOrder(
			relations.NewFilter(
				relations.NewAlias(
					relations.NewProjection(
						relations.NewJoin(
							relations.NewJoin(
								tableA,
								tableB,
								nil,
							),
							relations.NewLimit(tableC, 100),
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
				filters.NewEquals(
					expressions.NewNamed("", "id"),
					expressions.NewSum(
						expressions.NewNamed("", "repository_id"),
						expressions.NewNamed("", "repository_id"),
					),
				),
			),
			expressions.NewNamed("", "id"),
		),
		3,
	)

	rows, err := relations.ScanRows(relations.ScanContext{}, relation)
	if err != nil {
		panic(err.Error())
	}

	displayValues(rows)
}

func displayValues(rows shared.Rows) {
	allValues := make([][]string, 0, len(rows.Fields))
	for _, rowValues := range rows.Values {
		strValues := make([]string, 0, len(rowValues))
		for _, value := range rowValues {
			strValues = append(strValues, fmt.Sprintf("%v", value))
		}

		allValues = append(allValues, strValues)
	}

	columnWidths := make([]int, len(rows.Fields))
	for i, field := range rows.Fields {
		columnWidths[i] = len(field.Name)

		for _, values := range allValues {
			if columnWidths[i] < len(values[i]) {
				columnWidths[i] = len(values[i])
			}
		}
	}

	for i, field := range rows.Fields {
		if i != 0 {
			fmt.Printf(" | ")
		}
		fmt.Printf(fmt.Sprintf("%% %ds", columnWidths[i]/2), field.Name)
	}
	fmt.Printf("\n")

	for i := range rows.Fields {
		if i != 0 {
			fmt.Printf("-+-")
		}
		fmt.Printf(strings.Repeat("-", columnWidths[i]))
	}
	fmt.Printf("\n")

	for _, values := range allValues {
		for i, value := range values {
			if i != 0 {
				fmt.Printf(" | ")
			}

			fmt.Printf(fmt.Sprintf("%% %ds", columnWidths[i]), value)
		}

		fmt.Printf("\n")
	}

	fmt.Printf("(%d rows)\n", len(allValues))
}
