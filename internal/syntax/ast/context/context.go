package context

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type ResolveContext struct {
	Tables TableGetter
	Scopes []Scope
}

type Scope struct {
	fields []fields.ResolvedField
}

type TableGetter interface {
	Get(name string) (impls.Table, bool)
}

func (rc *ResolveContext) WithScope(f func() error) error {
	rc.PushScope()
	defer rc.PopScope()

	return f()
}

func (rc *ResolveContext) PushScope() {
	rc.Scopes = append(rc.Scopes, Scope{})
}

func (rc *ResolveContext) PopScope() {
	rc.Scopes = rc.Scopes[:len(rc.Scopes)-1]
}

func (rc *ResolveContext) CurrentScope() *Scope {
	if len(rc.Scopes) == 0 {
		panic("no scopes in context")
	}

	return &rc.Scopes[len(rc.Scopes)-1]
}

// TODO - actually resolve
func (rc *ResolveContext) Bind(fields ...fields.ResolvedField) {
	current := rc.CurrentScope()
	current.fields = append(current.fields, fields...)
}

func (rc *ResolveContext) Lookup(relationName, name string) (fields.ResolvedField, error) {
	for i := len(rc.Scopes) - 1; i >= 0; i-- {
		candidates := []fields.ResolvedField{}
		for _, field := range rc.Scopes[i].fields {
			if (field.RelationName() == relationName || relationName == "") && field.Name() == name {
				candidates = append(candidates, field)
			}
		}

		if len(candidates) > 1 {
			return fields.ResolvedField{}, fmt.Errorf("ambiguous field %q", name)
		}

		if len(candidates) == 1 {
			return candidates[0], nil
		}
	}

	return fields.ResolvedField{}, fmt.Errorf("unknown field %q", name)
}

func (rc *ResolveContext) ResolveExpression(expr impls.Expression) (impls.Expression, error) {
	return expr.Map(func(e impls.Expression) (impls.Expression, error) {
		if named, ok := e.(expressions.NamedExpression); ok {
			field, err := rc.Lookup(named.Field().RelationName(), named.Field().Name())
			if err != nil {
				return nil, err
			}

			return expressions.NewNamed(field), nil
		}

		return e, nil
	})
}
