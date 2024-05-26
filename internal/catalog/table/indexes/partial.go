package indexes

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

type partialIndex[O types.ScanOptions] struct {
	types.Index[O]
	condition types.Expression
}

var _ types.Index[types.ScanOptions] = &partialIndex[types.ScanOptions]{}

func NewPartialIndex[O types.ScanOptions](index types.Index[O], condition types.Expression) *partialIndex[O] {
	return &partialIndex[O]{
		Index:     index,
		condition: condition,
	}
}

func (i *partialIndex[O]) Unwrap() types.BaseIndex {
	return i.Index
}

func (i *partialIndex[O]) Filter() types.Expression {
	return i.condition
}

func (i *partialIndex[O]) Insert(row shared.Row) error {
	if i.condition != nil {
		if ok, err := shared.ValueAs[bool](i.condition.ValueFrom(types.EmptyContext, row)); err != nil {
			return err
		} else if ok == nil || !*ok {
			return nil
		}
	}

	return i.Index.Insert(row)
}

func (i *partialIndex[O]) Delete(row shared.Row) error {
	if i.condition != nil {
		if ok, err := shared.ValueAs[bool](i.condition.ValueFrom(types.EmptyContext, row)); err != nil {
			return err
		} else if ok == nil || !*ok {
			return nil
		}
	}

	return i.Index.Delete(row)
}

func matchesPartial(index types.BaseIndex, filterExpression types.Expression) bool {
	indexFilter := index.Filter()
	if indexFilter == nil {
		return true
	}

	if filterExpression == nil {
		return false
	}

	// TODO - need to do a more tight "subsumes" check
	for _, v := range expressions.Conjunctions(indexFilter) {
		if diff := expressions.FilterDifference(v, filterExpression); diff != nil && len(expressions.Conjunctions(diff)) >= len(expressions.Conjunctions(filterExpression)) {
			return false
		}
	}

	return true
}
