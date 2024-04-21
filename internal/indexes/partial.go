package indexes

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries/filter"
	"github.com/efritz/gostgres/internal/shared"
)

type partialIndex[O ScanOptions] struct {
	Index[O]
	condition expressions.Expression
}

var _ Index[ScanOptions] = &partialIndex[ScanOptions]{}

func NewPartialIndex[O ScanOptions](index Index[O], condition expressions.Expression) *partialIndex[O] {
	return &partialIndex[O]{
		Index:     index,
		condition: condition,
	}
}

func (i *partialIndex[O]) Unwrap() BaseIndex {
	return i.Index
}

func (i *partialIndex[O]) Filter() expressions.Expression {
	return i.condition
}

func (i *partialIndex[O]) Insert(row shared.Row) error {
	if i.condition != nil {
		if ok, err := shared.ValueAs[bool](i.condition.ValueFrom(row)); err != nil {
			return err
		} else if ok == nil || !*ok {
			return nil
		}
	}

	return i.Index.Insert(row)
}

func (i *partialIndex[O]) Delete(row shared.Row) error {
	if i.condition != nil {
		if ok, err := shared.ValueAs[bool](i.condition.ValueFrom(row)); err != nil {
			return err
		} else if ok == nil || !*ok {
			return nil
		}
	}

	return i.Index.Delete(row)
}

func matchesPartial(index BaseIndex, filterExpression expressions.Expression) bool {
	indexFilter := index.Filter()
	if indexFilter == nil {
		return true
	}

	if filterExpression == nil {
		return false
	}

	// TODO - need to do a more tight "subsumes" check
	for _, v := range indexFilter.Conjunctions() {
		if diff := filter.FilterDifference(v, filterExpression); diff != nil && len(diff.Conjunctions()) >= len(filterExpression.Conjunctions()) {
			return false
		}
	}

	return true
}
