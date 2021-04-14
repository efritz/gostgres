package relations

import (
	"bytes"
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type offsetRelation struct {
	Relation
	offset int
}

var _ Relation = &offsetRelation{}

func NewOffset(table Relation, offset int) Relation {
	return &offsetRelation{
		Relation: table,
		offset:   offset,
	}
}

func (r *offsetRelation) Serialize(buf *bytes.Buffer, indentationLevel int) {
	buf.WriteString(fmt.Sprintf("%soffset %d\n", indent(indentationLevel), r.offset))
	r.Relation.Serialize(buf, indentationLevel+1)
}

func (r *offsetRelation) Scan(visitor VisitorFunc) error {
	offset := r.offset
	return r.Relation.Scan(func(row shared.Row) (bool, error) {
		offset--
		if offset >= 0 {
			return true, nil
		}

		return visitor(row)
	})
}
