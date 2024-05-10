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
//	| `SEQUENCE` createSequence
//	| [ `UNIQUE` ] `INDEX` createIndex
func (p *parser) parseCreate(token tokens.Token) (queries.Query, error) {
	if p.advanceIf(isType(tokens.TokenTypeTable)) {
		return p.parseCreateTable()
	}

	if p.advanceIf(isType(tokens.TokenTypeSequence)) {
		return p.parseCreateSequence()
	}

	unique := false
	if p.advanceIf(isType(tokens.TokenTypeUnique)) {
		unique = true
	}

	if p.advanceIf(isType(tokens.TokenTypeIndex)) {
		return p.parseCreateIndex(unique)
	} else if unique {
		return nil, fmt.Errorf("expected create index statement (near %s)", p.current().Text)
	}

	return nil, fmt.Errorf("expected create statement (near %s)", p.current().Text)
}

// createTable := name `(` [ column [, ...] ] `)`
func (p *parser) parseCreateTable() (queries.Query, error) {
	name, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
	if err != nil {
		return nil, err
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeLeftParen)); err != nil {
		return nil, err
	}

	columns := []columnDescription{}
	if !p.advanceIf(isType(tokens.TokenTypeRightParen)) {
		for {
			column, err := p.parseColumn(name.Text)
			if err != nil {
				return nil, err
			}

			columns = append(columns, column)

			if !p.advanceIf(isType(tokens.TokenTypeComma)) {
				break
			}
		}

		if _, err := p.mustAdvance(isType(tokens.TokenTypeRightParen)); err != nil {
			return nil, err
		}
	}

	var fields []shared.Field
	var sequences []ddl.DDLQuery
	var constraints []ddl.DDLQuery
	for _, column := range columns {
		fields = append(fields, column.field)
		sequences = append(sequences, column.sequences...)
		constraints = append(constraints, column.constraints...)
	}

	return ddl.NewSet(append(append(sequences, ddl.NewCreateTable(name.Text, fields)), constraints...)), nil
}

type columnDescription struct {
	field       shared.Field
	sequences   []ddl.DDLQuery
	constraints []ddl.DDLQuery
}

// column := columnName dataType [( NOT NULL )] [ CHECK ( expression )] [ REFERENCES reftable `(` columnName `)` ] [ NOT NULL ] [ DEFAULT expression ]
//
// TODO: if not exists
// TODO: unique, primary key, reference column constraints
// TODO: named column constraints
// TODO: table constraints
func (p *parser) parseColumn(tableName string) (columnDescription, error) {
	name, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
	if err != nil {
		return columnDescription{}, err
	}

	dataType, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
	if err != nil {
		return columnDescription{}, err
	}

	var typ shared.Type
	var sequences []ddl.DDLQuery
	var constraints []ddl.DDLQuery
	var defaultExpression expressions.Expression

	makeSequence := func() {
		sequenceName := fmt.Sprintf("%s_%s_seq", tableName, name.Text)
		sequences = append(sequences, ddl.NewCreateSequence(sequenceName, typ))
		defaultExpression = expressions.NewFunction("nextval", []expressions.Expression{expressions.NewConstant(sequenceName)})
	}

	switch strings.ToLower(dataType.Text) {
	case "smallserial":
		typ = shared.TypeSmallInteger
		makeSequence()
	case "serial":
		typ = shared.TypeInteger
		makeSequence()
	case "bigserial":
		typ = shared.TypeBigInteger
		makeSequence()

	default:
		typ, err = p.parseBasicType(dataType.Text)
		if err != nil {
			return columnDescription{}, err
		}
	}

	for {
		if p.advanceIf(isType(tokens.TokenTypeNotNull)) {
			typ = typ.NonNullable()
			continue
		}

		if p.advanceIf(isType(tokens.TokenTypePrimaryKey)) {
			typ = typ.NonNullable()
			constraints = append(constraints, ddl.NewCreatePrimaryKeyConstraint(
				fmt.Sprintf("%s_pkey", tableName),
				tableName,
				[]string{name.Text},
			))
			continue
		}

		if p.advanceIf(isType(tokens.TokenTypeReferences)) {
			refTable, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
			if err != nil {
				return columnDescription{}, err
			}

			// TODO - refcolumn name should be optional, but we need a way
			// to determine the primary key of the target table otherwise.
			if _, err := p.mustAdvance(isType(tokens.TokenTypeLeftParen)); err != nil {
				return columnDescription{}, err
			}

			refColumn, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
			if err != nil {
				return columnDescription{}, err
			}

			if _, err := p.mustAdvance(isType(tokens.TokenTypeRightParen)); err != nil {
				return columnDescription{}, err
			}

			constraints = append(constraints, ddl.NewCreateForeignKeyConstraint(
				fmt.Sprintf("%s_%s_fkey", tableName, name.Text),
				tableName,
				[]string{name.Text},
				refTable.Text,
				[]string{refColumn.Text},
			))
			continue
		}

		if p.advanceIf(isType(tokens.TokenTypeCheck)) {
			token, err := p.mustAdvance(isType(tokens.TokenTypeLeftParen))
			if err != nil {
				return columnDescription{}, err
			}

			expr, err := p.parseParenthesizedExpression(token)
			if err != nil {
				return columnDescription{}, err
			}

			constraints = append(constraints, ddl.NewCreateCheckConstraint(
				fmt.Sprintf("%s_%s_check", tableName, name.Text),
				tableName,
				expr,
			))
			continue
		}

		if p.advanceIf(isType(tokens.TokenTypeDefault)) {
			expression, err := p.parseExpression(0)
			if err != nil {
				return columnDescription{}, err
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

	return columnDescription{
		field:       field,
		sequences:   sequences,
		constraints: constraints,
	}, nil
}

func (p *parser) parseBasicType(name string) (shared.Type, error) {

	var typ shared.Type
	switch strings.ToLower(name) {
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
		// TODO - use multi-phrase keyword
	case "double":
		if !p.advanceIf(isIdent("precision")) {
			return shared.Type{}, fmt.Errorf("unknown type %q", "double")
		}
		typ = shared.TypeNullableDoublePrecision
	case "numeric":
		typ = shared.TypeNullableNumeric
	case "boolean":
		typ = shared.TypeNullableBool
		// TODO - use multi-phrase keyword(s)
	case "timestamp":
		if !p.advanceIf(isIdent("with"), isIdent("time"), isIdent("zone")) {
			return shared.Type{}, fmt.Errorf("unknown type %q", "timestamp")
		}
		typ = shared.TypeNullableTimestampTz
	default:
		return shared.Type{}, fmt.Errorf("unknown type %s", name)
	}

	return typ, nil
}

// createSequence := name [ `AS` datatype ]
//
// TODO - increment, min/max value, start, cycle
func (p *parser) parseCreateSequence() (queries.Query, error) {
	name, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
	if err != nil {
		return nil, err
	}

	typ := shared.TypeBigInteger
	if p.advanceIf(isType(tokens.TokenTypeAs)) {
		dataType, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
		if err != nil {
			return nil, err
		}

		typ, err = p.parseBasicType(dataType.Text)
		if err != nil {
			return nil, err
		}
	}

	return ddl.NewCreateSequence(name.Text, typ), nil
}

// createIndex := name `ON` tableName [ `USING` methodName ] `(` expression [ `ASC` | `DESC` ] [, ...] `)` [ `WHERE` predicate ]
//
// TODO: if not exists
// TODO: concurrently
// TODO: NULLS FIRST | LAST
// TODO: include
// TODO: nulls distinct
func (p *parser) parseCreateIndex(unique bool) (queries.Query, error) {
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

// alter := `TABLE` tableName  alterTable
func (p *parser) parseAlter(token tokens.Token) (queries.Query, error) {
	if p.advanceIf(isType(tokens.TokenTypeTable)) {
		name, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
		if err != nil {
			return nil, err
		}

		return p.parseAlterTable(name.Text)
	}

	return nil, fmt.Errorf("expected alter statement (near %s)", p.current().Text)
}

// alterTable := `ADD` `CONSTRAINT` constraintName constraint
//
// constraint := `PRIMARY` `KEY` `(` columnName [ , ... ] `)`
//
//	| `FOREIGN` `KEY` `(` columnName [ , ... ] `)` `REFERENCES` refTable `(` refColumn [ , ... ] `)`
//	| `CHECK` `(` expr `)`
func (p *parser) parseAlterTable(tableName string) (queries.Query, error) {
	if p.advanceIf(isType(tokens.TokenTypeAdd), isType(tokens.TokenTypeConstraint)) {
		name, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
		if err != nil {
			return nil, err
		}

		if p.advanceIf(isType(tokens.TokenTypePrimaryKey)) {
			columnNames, err := p.mustParseColumnNames()
			if err != nil {
				return nil, err
			}

			// TODO - also ensure column names are not-null
			return ddl.NewCreatePrimaryKeyConstraint(name.Text, tableName, columnNames), nil
		}

		if p.advanceIf(isType(tokens.TokenTypeForeignKey)) {
			columnNames, err := p.mustParseColumnNames()
			if err != nil {
				return nil, err
			}

			if _, err := p.mustAdvance(isType(tokens.TokenTypeReferences)); err != nil {
				return nil, err
			}

			refTable, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
			if err != nil {
				return nil, err
			}

			refColumnNames, err := p.mustParseColumnNames()
			if err != nil {
				return nil, err
			}

			return ddl.NewCreateForeignKeyConstraint(name.Text, tableName, columnNames, refTable.Text, refColumnNames), nil
		}

		if p.advanceIf(isType(tokens.TokenTypeCheck)) {
			token, err := p.mustAdvance(isType(tokens.TokenTypeLeftParen))
			if err != nil {
				return nil, err
			}

			expr, err := p.parseParenthesizedExpression(token)
			if err != nil {
				return nil, err
			}

			return ddl.NewCreateCheckConstraint(name.Text, tableName, expr), nil
		}

		return nil, fmt.Errorf("expected add constraint statement (near %s)", p.current().Text)
	}

	return nil, fmt.Errorf("expected alter table statement (near %s)", p.current().Text)
}
