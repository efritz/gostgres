package indexes

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

type partialIndex[O impls.ScanOptions] struct {
	impls.Index[O]
	condition impls.Expression
}

var _ impls.Index[impls.ScanOptions] = &partialIndex[impls.ScanOptions]{}

func NewPartialIndex[O impls.ScanOptions](index impls.Index[O], condition impls.Expression) *partialIndex[O] {
	return &partialIndex[O]{
		Index:     index,
		condition: condition,
	}
}

func (i *partialIndex[O]) Unwrap() impls.BaseIndex {
	return i.Index
}

func (i *partialIndex[O]) Filter() impls.Expression {
	return i.condition
}

func (i *partialIndex[O]) Insert(row rows.Row) error {
	if i.condition != nil {
		if ok, err := types.ValueAs[bool](i.condition.ValueFrom(impls.EmptyContext, row)); err != nil {
			return err
		} else if ok == nil || !*ok {
			return nil
		}
	}

	return i.Index.Insert(row)
}

func (i *partialIndex[O]) Delete(row rows.Row) error {
	if i.condition != nil {
		if ok, err := types.ValueAs[bool](i.condition.ValueFrom(impls.EmptyContext, row)); err != nil {
			return err
		} else if ok == nil || !*ok {
			return nil
		}
	}

	return i.Index.Delete(row)
}

func matchesPartial(index impls.BaseIndex, filterExpression impls.Expression) bool {
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
