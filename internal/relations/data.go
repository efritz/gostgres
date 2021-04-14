package relations

import (
	"bytes"
	"fmt"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type dataRelation struct {
	name string
	rows shared.Rows

	filter expressions.Expression
	order  expressions.Expression
	// TODO - track fields for index only scans as well
}

var _ Relation = &dataRelation{}

func NewData(name string, rows shared.Rows) Relation {
	return &dataRelation{
		name: name,
		rows: rows,
	}
}

func (r *dataRelation) Name() string           { return r.name }
func (r *dataRelation) Fields() []shared.Field { return copyFields(r.rows.Fields) }

func (r *dataRelation) Serialize(buf *bytes.Buffer, indentationLevel int) {
	scanType := "seq"
	if r.filter != nil {
		scanType = "index"
	}

	buf.WriteString(fmt.Sprintf("%s%s scan over %s\n", indent(indentationLevel), scanType, r.name))
	if r.filter != nil {
		buf.WriteString(fmt.Sprintf("%sfilter: %s\n", indent(indentationLevel+1), r.filter))
	}
	if r.order != nil {
		// TODO - how to serialize this?
	}
}

func (r *dataRelation) Optimize() {
	// no-op
}

func (r *dataRelation) SinkFilter(filter expressions.Expression) bool {
	// TODO - only accept if we can
	r.filter = filter
	return true
}

func (r *dataRelation) Scan(visitor VisitorFunc) error {
	indexes, err := findIndexIterationOrder(r.order, r.rows)
	if err != nil {
		return err
	}

	for _, i := range indexes {
		row := r.rows.Row(i)

		if r.filter != nil {
			if ok, err := expressions.Bool(r.filter).ValueFrom(row); err != nil {
				return err
			} else if !ok {
				continue
			}
		}

		if ok, err := visitor(row); err != nil {
			return err
		} else if !ok {
			break
		}
	}

	return nil
}
