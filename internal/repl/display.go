package repl

import (
	"fmt"
	"io"
	"strings"

	"github.com/efritz/gostgres/internal/nodes"
	"github.com/efritz/gostgres/internal/shared"
)

func serializePlan(w io.Writer, node nodes.Node) {
	node.Serialize(w, 0)
}

func serializeRows(w io.Writer, rows shared.Rows) {
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

	fmt.Fprintf(w, " ")
	for i, field := range rows.Fields {
		if i != 0 {
			fmt.Fprintf(w, " | ")
		}

		name := field.Name
		if len(field.Name) < columnWidths[i] {
			name += strings.Repeat(" ", (columnWidths[i]-len(field.Name))/2)
		}
		fmt.Fprintf(w, fmt.Sprintf("%% %ds", columnWidths[i]), name)
	}
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "-")
	for i := range rows.Fields {
		if i != 0 {
			fmt.Fprintf(w, "-+-")
		}
		fmt.Fprintf(w, strings.Repeat("-", columnWidths[i]))
	}
	fmt.Fprintf(w, "-")
	fmt.Fprintf(w, "\n")

	for _, values := range allValues {
		fmt.Fprintf(w, " ")

		for i, value := range values {
			if i != 0 {
				fmt.Fprintf(w, " | ")
			}

			fmt.Fprintf(w, fmt.Sprintf("%% %ds", columnWidths[i]), value)
		}

		fmt.Fprintf(w, "\n")
	}

	fmt.Fprintf(w, "(%d rows)\n", len(allValues))
}
