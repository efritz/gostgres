package ast

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/impls"
)

type ResolutionContext struct {
	Tables TableGetter
	Scopes []Scope
}

func (rc *ResolutionContext) PushScope() {
	rc.Scopes = append(rc.Scopes, Scope{})
}

func (rc *ResolutionContext) PopScope() {
	rc.Scopes = rc.Scopes[:len(rc.Scopes)-1]
}

func (rc *ResolutionContext) CurrentScope() *Scope {
	if len(rc.Scopes) == 0 {
		panic("no scopes in context")
	}

	return &rc.Scopes[len(rc.Scopes)-1]
}

func (rc *ResolutionContext) AddTableAlias(alias, relationName string) {
	scope := rc.CurrentScope()
	scope.TableAliases = append(scope.TableAliases, TableAliasInfo{
		Alias:        alias,
		RelationName: relationName,
	})
}

func (rc *ResolutionContext) AddColumnAlias(alias, relationName, columnName string) {
	scope := rc.CurrentScope()
	scope.ColumnAliases = append(scope.ColumnAliases, ColumnAliasInfo{
		Alias:        alias,
		RelationName: relationName,
		ColumnName:   columnName,
	})
}

func (rc *ResolutionContext) GetTableAlias(alias string) (string, error) {
	for i := len(rc.Scopes) - 1; i >= 0; i-- {
		scope := rc.Scopes[i]

		if info, err := scope.GetTableAlias(alias); !isNotFound(err) {
			return info.RelationName, err
		}
	}

	return "", notFoundError{"table", alias}
}

func (rc *ResolutionContext) GetColumnAlias(relationName, alias string) (string, string, error) {
	for i := len(rc.Scopes) - 1; i >= 0; i-- {
		scope := rc.Scopes[i]

		if info, err := scope.GetColumnAlias(relationName, alias); !isNotFound(err) {
			return info.RelationName, info.ColumnName, err
		}
	}

	return "", "", notFoundError{"column", alias}
}

type Scope struct {
	TableAliases  []TableAliasInfo
	ColumnAliases []ColumnAliasInfo
}

func (s *Scope) GetTableAlias(alias string) (TableAliasInfo, error) {
	for _, tableAlias := range s.TableAliases {
		if tableAlias.Alias == alias {
			return tableAlias, nil
		}
	}

	return TableAliasInfo{}, notFoundError{"table", alias}
}

type notFoundError struct {
	typ   string
	alias string
}

func (e notFoundError) Error() string {
	return fmt.Sprintf("unknown %s alias %q", e.typ, e.alias)
}

func isNotFound(err error) bool {
	_, ok := err.(notFoundError)
	return ok
}

func (s *Scope) GetColumnAlias(relationName, alias string) (ColumnAliasInfo, error) {
	var found ColumnAliasInfo
	var foundCount int
	for _, columnAlias := range s.ColumnAliases {
		if columnAlias.Alias == alias && (relationName == "" || columnAlias.RelationName == relationName) {
			found = columnAlias
			foundCount++
		}
	}

	if foundCount == 0 {
		return ColumnAliasInfo{}, notFoundError{"column", alias}
	}

	if foundCount > 1 {
		return ColumnAliasInfo{}, fmt.Errorf("ambiguous column alias %q", alias)
	}

	return found, nil
}

type TableAliasInfo struct {
	Alias        string
	RelationName string // TODO
}

type ColumnAliasInfo struct {
	Alias        string
	RelationName string // TODO
	ColumnName   string // TODO
}

// type ExpressionAliasInfo struct {
// 	// TODO
// }

//
//
//

type BuildContext struct {
	Tables TableGetter // TODO - necessary if used in resolution pass?
}

type TableGetter interface {
	Get(name string) (impls.Table, bool)
}
