package expressions

import (
	"bytes"
	"fmt"

	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type Query interface {
	Serialize(w serialization.IndentWriter)
	Optimize()
	Scanner(ctx impls.Context) (scan.RowScanner, error)
}

//
//
//

type existsSubqueryExpression struct {
	subquery Query
}

func NewExistsSubqueryExpression(subquery Query) impls.Expression {
	return existsSubqueryExpression{
		subquery: subquery,
	}
}

func (e existsSubqueryExpression) String() string                    { return "" }    // TODO
func (e existsSubqueryExpression) Equal(other impls.Expression) bool { return false } // TODO
func (e existsSubqueryExpression) Children() []impls.Expression      { return nil }   // TODO
func (e existsSubqueryExpression) Fold() impls.Expression            { return e }     // TODO
func (e existsSubqueryExpression) Map(f func(impls.Expression) impls.Expression) impls.Expression {
	return e
} // TODO

func (e existsSubqueryExpression) ValueFrom(ctx impls.Context, row rows.Row) (any, error) {
	// TODO - want to pull this out and do optimization before the enclosing query is executed
	// TODO - need to pull out serialization into separate parts in the explain output as well
	ctx.Log("Optimizing subquery")
	e.subquery.Optimize()

	var buf bytes.Buffer
	w := serialization.NewIndentWriter(&buf)
	e.subquery.Serialize(w)
	ctx.Log("%s", buf.String())

	ctx.Log("Preparing subquery over %s", row)

	s, err := e.subquery.Scanner(ctx.WithOuterRow(row))
	if err != nil {
		return nil, err
	}

	ctx.Log("Scanning subquery")

	if _, err := s.Scan(); err != nil {
		if err != scan.ErrNoRows {
			return nil, err
		}

		return false, err
	}

	return true, nil
}

//
//
//

type anySubqueryExpression struct {
	expression impls.Expression
	op         string // TODO
	subquery   Query
}

func NewAnySubqueryExpression(expression impls.Expression, op string, subquery Query) impls.Expression {
	return anySubqueryExpression{
		expression: expression,
		op:         op,
		subquery:   subquery,
	}
}

func (e anySubqueryExpression) String() string                    { return "" }    // TODO
func (e anySubqueryExpression) Equal(other impls.Expression) bool { return false } // TODO
func (e anySubqueryExpression) Children() []impls.Expression      { return nil }   // TODO
func (e anySubqueryExpression) Fold() impls.Expression            { return e }     // TODO
func (e anySubqueryExpression) Map(f func(impls.Expression) impls.Expression) impls.Expression {
	return e
} // TODO

func (e anySubqueryExpression) ValueFrom(ctx impls.Context, row rows.Row) (any, error) {
	return nil, fmt.Errorf("unimplemented") // TODO
}

//
//
//

type allSubqueryExpression struct {
	expression impls.Expression
	op         string // TODO
	subquery   Query
}

func NewAllSubqueryExpression(expression impls.Expression, op string, subquery Query) impls.Expression {
	return allSubqueryExpression{
		expression: expression,
		op:         op,
		subquery:   subquery,
	}
}

func (e allSubqueryExpression) String() string                    { return "" }    // TODO
func (e allSubqueryExpression) Equal(other impls.Expression) bool { return false } // TODO
func (e allSubqueryExpression) Children() []impls.Expression      { return nil }   // TODO
func (e allSubqueryExpression) Fold() impls.Expression            { return e }     // TODO
func (e allSubqueryExpression) Map(f func(impls.Expression) impls.Expression) impls.Expression {
	return e
} // TODO

func (e allSubqueryExpression) ValueFrom(ctx impls.Context, row rows.Row) (any, error) {
	return nil, fmt.Errorf("unimplemented") // TODO
}

//
//
//

type opSubqueryExpression struct {
	expression impls.Expression
	op         string // TODO
	subquery   Query
}

func NewOpSubqueryExpression(expression impls.Expression, op string, subquery Query) impls.Expression {
	return opSubqueryExpression{
		expression: expression,
		op:         op,
		subquery:   subquery,
	}
}

func (e opSubqueryExpression) String() string                    { return "" }    // TODO
func (e opSubqueryExpression) Equal(other impls.Expression) bool { return false } // TODO
func (e opSubqueryExpression) Children() []impls.Expression      { return nil }   // TODO
func (e opSubqueryExpression) Fold() impls.Expression            { return e }     // TODO
func (e opSubqueryExpression) Map(f func(impls.Expression) impls.Expression) impls.Expression {
	return e
} // TODO

func (e opSubqueryExpression) ValueFrom(ctx impls.Context, row rows.Row) (any, error) {
	return nil, fmt.Errorf("unimplemented") // TODO
}
