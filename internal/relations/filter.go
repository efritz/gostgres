package relations

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type filterRelation struct {
	Relation
	filter expressions.Expression
}

var _ Relation = &filterRelation{}

func NewFilter(relation Relation, filter expressions.Expression) Relation {
	return &filterRelation{
		Relation: relation,
		filter:   filter,
	}
}

func (r *filterRelation) Serialize(w io.Writer, indentationLevel int) {
	if r.filter == nil {
		r.Relation.Serialize(w, indentationLevel)
		return
	}

	io.WriteString(w, fmt.Sprintf("%sfilter by %s\n", indent(indentationLevel), r.filter))
	r.Relation.Serialize(w, indentationLevel+1)
}

func (r *filterRelation) Optimize() {
	if r.filter != nil {
		r.filter = r.distributeFilter(r.filter.Fold())
	}

	r.Relation.Optimize()
}

func (r *filterRelation) distributeFilter(filter expressions.Expression) expressions.Expression {
	var conjunctions []expressions.Expression
	for _, expression := range filter.Conjunctions() {
		if !r.distributeExpression(expression) {
			conjunctions = append(conjunctions, expression)
		}
	}

	return combineConjunctions(conjunctions)
}

func (r *filterRelation) distributeExpression(expression expressions.Expression) bool {
	return r.Relation.PushDownFilter(expression)
}

func (r *filterRelation) PushDownFilter(filter expressions.Expression) bool {
	if r.filter != nil {
		filter = expressions.NewAnd(r.filter, filter)
	}

	r.filter = filter
	return true
}

func (r *filterRelation) Scan(visitor VisitorFunc) error {
	if r.filter == nil {
		return r.Relation.Scan(visitor)
	}

	return r.Relation.Scan(func(row shared.Row) (bool, error) {
		if ok, err := expressions.EnsureBool(r.filter.ValueFrom(row)); err != nil {
			return false, err
		} else if !ok {
			return true, nil
		}

		return visitor(row)
	})
}
