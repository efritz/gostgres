package serialization

import (
	"bytes"
	"io"

	"github.com/efritz/gostgres/internal/nodes"
)

func SerializePlanString(node nodes.Node) string {
	var buf bytes.Buffer
	SerializePlan(&buf, node)
	return buf.String()
}

func SerializePlan(w io.Writer, node nodes.Node) {
	node.Serialize(w, 0)
}
