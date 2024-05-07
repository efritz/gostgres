package parsing

import (
	"fmt"
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/queries/ddl"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// create := `TABLE` createTable
//
//	| [ `UNIQUE` ] `INDEX` createIndex
func (p *parser) parseCreate(token tokens.Token) (queries.Query, error) {
	if p.advanceIf(isType(tokens.TokenTypeTable)) {
		return p.parseCreateTable(token)
	}

	unique := false
	if p.advanceIf(isType(tokens.TokenTypeUnique)) {
		unique = true
	}

	if p.advanceIf(isType(tokens.TokenTypeIndex)) {
		return p.parseCreateIndex(token, unique)
	} else if unique {
		return nil, fmt.Errorf("expected create index statement (near %s)", p.current().Text)
	}

	return nil, fmt.Errorf("expected create statement (near %s)", p.current().Text)
}

// createTable := name `(` [ column [, ...] ] `)`
func (p *parser) parseCreateTable(token tokens.Token) (queries.Query, error) {
	name, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
	if err != nil {
		return nil, err
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeLeftParen)); err != nil {
		return nil, err
	}

	fields := []shared.Field{}
	if !p.advanceIf(isType(tokens.TokenTypeRightParen)) {
		for {
			field, err := p.parseColumn()
			if err != nil {
				return nil, err
			}

			fields = append(fields, field)

			if !p.advanceIf(isType(tokens.TokenTypeComma)) {
				break
			}
		}

		if _, err := p.mustAdvance(isType(tokens.TokenTypeRightParen)); err != nil {
			return nil, err
		}
	}

	return ddl.NewCreateTable(name.Text, fields), nil
}

// column := columnName dataType [( NOT NULL )]
//
// TODO: if not exists
// TODO: table constraints
// TODO: column constraints
func (p *parser) parseColumn() (shared.Field, error) {
	name, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
	if err != nil {
		return shared.Field{}, err
	}

	dataType, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
	if err != nil {
		return shared.Field{}, err
	}

	var typ shared.Type
	switch strings.ToLower(dataType.Text) {
	case "text":
		typ = shared.TypeNullableText
	case "smallint":
		typ = shared.TypeNullableSmallInteger
	case "integer":
		typ = shared.TypeNullableInteger
	case "bigint":
		typ = shared.TypeNullableBigInteger
	case "real":
		typ = shared.TypeNullableReal
	case "double":
		if !p.advanceIf(isIdent("precision")) {
			return shared.Field{}, fmt.Errorf("unknown type %q", "double")
		}
		typ = shared.TypeNullableDoublePrecision
	case "numeric":
		typ = shared.TypeNullableNumeric
	case "boolean":
		typ = shared.TypeNullableBool
	case "timestamp":
		if !p.advanceIf(isIdent("with"), isIdent("time"), isIdent("zone")) {
			return shared.Field{}, fmt.Errorf("unknown type %q", "timestamp")
		}
		typ = shared.TypeNullableTimestampTz
	default:
		return shared.Field{}, fmt.Errorf("unknown type %s", dataType.Text)
	}

	var defaultExpression expressions.Expression
	for {
		if p.advanceIf(isType(tokens.TokenTypeNotNull)) {
			typ = typ.NonNullable()
			continue
		}

		if p.advanceIf(isType(tokens.TokenTypeDefault)) {
			expression, err := p.parseExpression(0)
			if err != nil {
				return shared.Field{}, err
			}

			defaultExpression = expression
			continue
		}

		break
	}

	field := shared.NewField("", name.Text, typ)

	if defaultExpression != nil {
		field = field.WithDefault(func() any {
			// TODO - need to do this lazily, store as expressions
			value, err := defaultExpression.ValueFrom(expressions.EmptyContext, shared.Row{})
			if err != nil {
				panic(err.Error()) // TODO
			}
			return value
		})
	}

	return field, nil
}

// createIndex := name `ON` tableName [ `USING` methodName ] `(` expression [ `ASC` | `DESC` ] [, ...] `)` [ `WHERE` predicate ]
//
// TODO: if not exists
// TODO: concurrently
// TODO: NULLS FIRST | LAST
// TODO: include
// TODO: nulls distinct
func (p *parser) parseCreateIndex(token tokens.Token, unique bool) (queries.Query, error) {
	name, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
	if err != nil {
		return nil, err
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeOn)); err != nil {
		return nil, err
	}

	tableName, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
	if err != nil {
		return nil, err
	}

	method := "btree"
	if p.advanceIf(isType(tokens.TokenTypeUsing)) {
		methodToken, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
		if err != nil {
			return nil, err
		}

		method = methodToken.Text
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeLeftParen)); err != nil {
		return nil, err
	}

	var columnExpressions []expressions.ExpressionWithDirection
	if !p.advanceIf(isType(tokens.TokenTypeRightParen)) {
		for {
			expression, err := p.parseExpression(0)
			if err != nil {
				return nil, err
			}

			reverse := false
			if p.advanceIf(isType(tokens.TokenTypeAscending)) {
				// no-op
			} else if p.advanceIf(isType(tokens.TokenTypeDescending)) {
				reverse = true
			}

			columnExpressions = append(columnExpressions, expressions.ExpressionWithDirection{
				Expression: expression,
				Reverse:    reverse,
			})

			if !p.advanceIf(isType(tokens.TokenTypeComma)) {
				break
			}
		}

		if _, err := p.mustAdvance(isType(tokens.TokenTypeRightParen)); err != nil {
			return nil, err
		}
	}

	var where expressions.Expression
	if p.advanceIf(isType(tokens.TokenTypeWhere)) {
		whereExpression, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		where = whereExpression
	}

	return ddl.NewCreateIndex(name.Text, tableName.Text, method, unique, columnExpressions, where), nil
}
