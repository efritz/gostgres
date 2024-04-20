package serialization

import "strings"

const indentPerLevel = 4

func Indent(level int) string {
	return strings.Repeat(" ", level*indentPerLevel)
}
