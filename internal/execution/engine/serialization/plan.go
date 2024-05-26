package serialization

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Serializable interface {
	Serialize(w IndentWriter)
}

func SerializePlan(node Serializable) string {
	var buf bytes.Buffer
	node.Serialize(NewIndentWriter(&buf))
	return strings.TrimRightFunc(buf.String(), unicode.IsSpace)
}

type IndentWriter struct {
	w     io.Writer
	level int
}

func NewIndentWriter(w io.Writer) IndentWriter {
	return IndentWriter{w: w}
}

func (w IndentWriter) Indent() IndentWriter {
	return IndentWriter{
		w:     w.w,
		level: w.level + 1,
	}
}

const indentPerLevel = 4

func (w IndentWriter) WritefLine(format string, args ...any) {
	fmt.Fprintf(w.w, "%s%s\n", strings.Repeat(" ", w.level*indentPerLevel), fmt.Sprintf(format, args...))
}
