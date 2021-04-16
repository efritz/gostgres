package relations

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type dataRelation struct {
	name  string
	table *Table

	filter expressions.Expression
	order  OrderExpression
	// TODO - track fields for index only scans as well
}

var _ Relation = &dataRelation{}

func NewData(name string, table *Table) Relation {
	return &dataRelation{
		name:  name,
		table: table,
	}
}

func (r *dataRelation) Name() string {
	return r.name
}

func (r *dataRelation) Fields() []shared.Field {
	return updateRelationName(r.table.Fields(), r.name)
}

func (r *dataRelation) Serialize(w io.Writer, indentationLevel int) {
	scanType := "seq"
	if r.filter != nil {
		scanType = "index"
	}

	io.WriteString(w, fmt.Sprintf("%s%s scan over %s\n", indent(indentationLevel), scanType, r.name))
	if r.filter != nil {
		io.WriteString(w, fmt.Sprintf("%sfilter: %s\n", indent(indentationLevel+1), r.filter))
	}
	if r.order != nil {
		// TODO - how to serialize this?
	}
}

func (r *dataRelation) Optimize() {
	if r.filter != nil {
		r.filter = r.filter.Fold()
	}

	if r.order != nil {
		r.order = r.order.Fold()
	}
}

func (r *dataRelation) PushDownFilter(filter expressions.Expression) bool {
	if r.filter != nil {
		filter = expressions.NewAnd(r.filter, filter)
	}

	r.filter = filter
	return true
}

func (r *dataRelation) Scan(visitor VisitorFunc) error {
	indexes, err := findIndexIterationOrder(r.order, r.table.rows)
	if err != nil {
		return err
	}

	for _, i := range indexes {
		row := r.table.rows.Row(i)

		if r.filter != nil {
			if ok, err := expressions.EnsureBool(r.filter.ValueFrom(row)); err != nil {
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
