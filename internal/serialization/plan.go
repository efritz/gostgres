package serialization

import (
	"bytes"

	"github.com/efritz/gostgres/internal/queries"
)

func SerializePlan(node queries.Node) string {
	var buf bytes.Buffer
	node.Serialize(&buf, 0)
	return buf.String()
}
