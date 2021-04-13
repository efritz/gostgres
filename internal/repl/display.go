package repl

import (
	"fmt"
	"strings"

	"github.com/efritz/gostgres/internal/shared"
)

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
