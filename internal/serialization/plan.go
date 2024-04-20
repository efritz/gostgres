package serialization

import (
	"bytes"
	"io"

	"github.com/efritz/gostgres/internal/queries"
)

func SerializePlanString(node queries.Node) string {
	var buf bytes.Buffer
	SerializePlan(&buf, node)
	return buf.String()
}

func SerializePlan(w io.Writer, node queries.Node) {
	node.Serialize(w, 0)
}
