package serialization

import (
	"fmt"
	"strings"

	"github.com/efritz/gostgres/internal/shared"
)

func SerializeRows(rows shared.Rows) string {
	var b strings.Builder
	serialized := serializeRows(rows)
	spacesBuffer := strings.Repeat(" ", max(serialized.maxFieldWidth, serialized.maxValueWidth))
	dashesBuffer := strings.Repeat("-", max(serialized.maxFieldWidth, serialized.maxValueWidth))

	// Write table header
	b.WriteRune(' ')
	for i, c := range serialized.columns {
		if i != 0 {
			b.WriteString(" | ")
		}

		// Center column name
		name := c.field.Name()
		padding := c.maxWidthWithFieldName - len(name)
		b.WriteString(spacesBuffer[:padding/2])
		b.WriteString(name)
		b.WriteString(spacesBuffer[:padding-padding/2])
	}
	b.WriteRune('\n')

	// Write table header separator
	b.WriteRune('-')
	for i, c := range serialized.columns {
		if i != 0 {
			b.WriteString("-+-")
		}
		b.WriteString(dashesBuffer[:c.maxWidthWithFieldName])
	}
	b.WriteRune('-')
	b.WriteRune('\n')

	// Write each row
	for i := range serialized.columns[0].values {
		b.WriteRune(' ')
		for j, c := range serialized.columns {
			if j != 0 {
				b.WriteString(" | ")
			}

			if c.field.Type().IsNumber() {
				// Right-align numeric types
				b.WriteString(spacesBuffer[:c.maxWidthWithFieldName-len(c.values[i])])
				b.WriteString(c.values[i])
			} else {
				// Left-align all other types
				for i, line := range strings.Split(c.values[i], "\n") {
					if i != 0 {
						b.WriteRune('\n')
					}

					b.WriteString(line)
					b.WriteString(spacesBuffer[:c.maxWidthWithFieldName-len(line)])
				}
			}
		}
		b.WriteRune('\n')
	}

	// Write row count trailer
	b.WriteString(fmt.Sprintf("(%d rows)\n", len(serialized.columns[0].values)))
	return b.String()
}

func SerializeRowsExpanded(rows shared.Rows) string {
	var b strings.Builder
	serialized := serializeRows(rows)
	maxLineWidth := serialized.maxFieldWidth + serialized.maxValueWidth + 3
	spacesBuffer := strings.Repeat(" ", serialized.maxFieldWidth)
	dashesBuffer := strings.Repeat("-", maxLineWidth-13) // 13 for `-[ RECORD x ]-`

	for i := 0; i < serialized.numRows; i++ {
		recordIndex := fmt.Sprintf("%d", i+1)
		b.WriteString("-[ RECORD ")
		b.WriteString(recordIndex)
		b.WriteString(" ]-")
		b.WriteString(dashesBuffer[:len(dashesBuffer)-len(recordIndex)])
		b.WriteRune('\n')

		for _, c := range serialized.columns {
			name := c.field.Name()
			b.WriteString(name)
			b.WriteString(spacesBuffer[:serialized.maxFieldWidth-len(name)])
			b.WriteString(" | ")

			for i, line := range strings.Split(c.values[i], "\n") {
				if i != 0 {
					b.WriteRune('\n')
					b.WriteString(spacesBuffer[:serialized.maxFieldWidth])
					b.WriteString(" | ")
				}

				b.WriteString(line)
			}
			b.WriteRune('\n')
		}
	}

	return b.String()
}

type serializedRows struct {
	columns       []serializedColumn
	numRows       int
	maxFieldWidth int // maximum length of all field names
	maxValueWidth int // maximum length of all serialized values
}

type serializedColumn struct {
	field                 shared.Field
	values                []string
	maxValueWidth         int // maximum length of serialized value in this column
	maxWidthWithFieldName int // max(maxValueWidth, len(field.Name()))
}

func serializeRows(rows shared.Rows) serializedRows {
	var columns []serializedColumn
	for _, field := range rows.Fields {
		columns = append(columns, serializedColumn{field: field})
	}

	for _, rowValues := range rows.Values {
		for i, value := range rowValues {
			strValue := ""
			switch value {
			case nil:
				strValue = "[NULL]"
			case true:
				strValue = "t"
			case false:
				strValue = "f"
			default:
				strValue = fmt.Sprintf("%v", value)
			}

			columns[i].values = append(columns[i].values, strValue)
			columns[i].maxValueWidth = max(columns[i].maxValueWidth, longestLine(strValue))
		}
	}

	maxFieldWidth := 0
	maxValueWidth := 0
	for i, c := range columns {
		nameLen := len(c.field.Name())
		maxFieldWidth = max(maxFieldWidth, nameLen)
		maxValueWidth = max(maxValueWidth, c.maxValueWidth)
		columns[i].maxWidthWithFieldName = max(c.maxValueWidth, nameLen)
	}

	return serializedRows{
		columns:       columns,
		numRows:       rows.Size(),
		maxFieldWidth: maxFieldWidth,
		maxValueWidth: maxValueWidth,
	}
}

func longestLine(text string) (maxLine int) {
	for _, line := range strings.Split(text, "\n") {
		maxLine = max(maxLine, len(line))
	}

	return maxLine
}
