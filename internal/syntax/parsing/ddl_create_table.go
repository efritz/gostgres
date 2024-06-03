package parsing

import (
	"fmt"
	"strings"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries/ddl"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/types"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// createTableTail := ident `(` [ columnDescription [, ...] ] `)`
func (p *parser) parseCreateTable() (Query, error) {
	name, err := p.parseIdent()
	if err != nil {
		return nil, err
	}

	columns, err := parseParenthesizedCommaSeparatedList(p, false, true, func() (columnDescription, error) {
		return p.parseColumnDescription(name)
	})
	if err != nil {
		return nil, err
	}

	var queries []ddl.DDLQuery
	for _, column := range columns {
		// Sequences must be created before tables that reference them
		queries = append(queries, column.sequences...)
	}

	var fields []impls.TableField
	for _, column := range columns {
		fields = append(fields, column.field)
	}
	queries = append(queries, ddl.NewCreateTable(name, fields))

	for _, column := range columns {
		// Constraints must be added after the table they reference
		queries = append(queries, column.constraints...)
	}

	// TODO - if not exists
	// TODO - table constraints
	return ddl.NewSet(queries), nil
}

type columnDescription struct {
	field       impls.TableField
	sequences   []ddl.DDLQuery
	constraints []ddl.DDLQuery
}

// columnDescription := ident columnType [ columnConstraint [...] ]
func (p *parser) parseColumnDescription(tableName string) (columnDescription, error) {
	name, err := p.parseIdent()
	if err != nil {
		return columnDescription{}, err
	}

	description := columnDescription{
		field:       impls.NewTableField("", name, types.TypeAny),
		sequences:   nil,
		constraints: nil,
	}

	if err := p.parseColumnType(name, tableName, &description); err != nil {
		return columnDescription{}, err
	}

	for {
		if matched, err := p.parseColumnConstraint(name, tableName, &description); err != nil {
			return columnDescription{}, err
		} else if !matched {
			break
		}
	}

	return description, nil
}

// columnType := basicType | `smallserial` | `serial` | `bigserial`
func (p *parser) parseColumnType(name string, tableName string, description *columnDescription) error {
	parseSequenceType := func() (types.Type, bool) {
		if p.peek(0).Type == tokens.TokenTypeIdent {
			switch strings.ToLower(p.peek(0).Text) {
			case "smallserial":
				p.advance()
				return types.TypeSmallInteger, true
			case "serial":
				p.advance()
				return types.TypeInteger, true
			case "bigserial":
				p.advance()
				return types.TypeBigInteger, true
			}
		}

		return types.TypeUnknown, false
	}

	if typ, ok := parseSequenceType(); ok {
		sequenceName := fmt.Sprintf("%s_%s_seq", tableName, name)
		description.sequences = append(description.sequences, ddl.NewCreateSequence(sequenceName, typ))
		description.field = description.field.WithType(typ)
		description.field = description.field.WithNonNullable()
		nextValue := expressions.NewFunction("nextval", []impls.Expression{expressions.NewConstant(sequenceName)})
		description.field = description.field.WithDefault(nextValue)
	} else {
		typ, err := p.parseBasicType()
		if err != nil {
			return err
		}

		description.field = description.field.WithType(typ)
	}

	return nil
}

func (p *parser) initColumnConstraintParsers() {
	p.columnConstraintParsers = columnConstraintParsers{
		tokens.TokenTypeNotNull:    p.parseNotNullColumnConstraint,
		tokens.TokenTypePrimaryKey: p.parsePrimaryKeyColumnConstraint,
		tokens.TokenTypeReferences: p.parseReferencesColumnConstraint,
		tokens.TokenTypeCheck:      p.parseCheckColumnConstraint,
		tokens.TokenTypeDefault:    p.parseDefaultColumnConstraint,
	}
}

// columnConstraint := ( `NOT NULL` ) | ( `PRIMARY KEY` ) | ( `CHECK` checkColumnConstraintTail ) | ( `REFERENCES` referencesColumnConstraintTail ) | ( `DEFAULT` defaultColumnConstraintTail )
func (p *parser) parseColumnConstraint(columnName, tableName string, description *columnDescription) (bool, error) {
	for tokenType, parser := range p.columnConstraintParsers {
		if p.advanceIf(isType(tokenType)) {
			return true, parser(columnName, tableName, description)
		}
	}

	// TODO - unique
	// TODO - named column constraints
	return false, nil
}

func (p *parser) parseNotNullColumnConstraint(columnName, tableName string, description *columnDescription) error {
	description.field = description.field.WithNonNullable()
	return nil
}

func (p *parser) parsePrimaryKeyColumnConstraint(columnName, tableName string, description *columnDescription) error {
	description.field = description.field.WithNonNullable()
	constraintName := fmt.Sprintf("%s_pkey", tableName)
	constraint := ddl.NewCreatePrimaryKeyConstraint(constraintName, tableName, []string{columnName})
	description.constraints = append(description.constraints, constraint)
	return nil
}

// referencesColumnConstraintTail := ident `(` ident `)`
func (p *parser) parseReferencesColumnConstraint(columnName, tableName string, description *columnDescription) error {
	refTable, err := p.parseIdent()
	if err != nil {
		return err
	}

	// TODO - refcolumn name should be optional, but we need a way
	// to determine the primary key of the target table otherwise.
	refColumnName, err := parseParenthesized(p, func() (string, error) {
		refColumn, err := p.parseIdent()
		if err != nil {
			return "", err
		}

		return refColumn, nil
	})
	if err != nil {
		return err
	}

	constraintName := fmt.Sprintf("%s_%s_fkey", tableName, columnName)
	constraint := ddl.NewCreateForeignKeyConstraint(
		constraintName,
		tableName,
		[]string{columnName},
		refTable,
		[]string{refColumnName},
	)
	description.constraints = append(description.constraints, constraint)
	return nil
}

// checkColumnConstraintTail := `(` expression `)`
func (p *parser) parseCheckColumnConstraint(columnName, tableName string, description *columnDescription) error {
	expr, err := parseParenthesized(p, p.parseRootExpression)
	if err != nil {
		return err
	}

	constraintName := fmt.Sprintf("%s_%s_check", tableName, columnName)
	constraint := ddl.NewCreateCheckConstraint(constraintName, tableName, expr)
	description.constraints = append(description.constraints, constraint)
	return nil
}

// defaultColumnConstraintTail := expression
func (p *parser) parseDefaultColumnConstraint(columnName, tableName string, description *columnDescription) error {
	expression, err := p.parseRootExpression()
	if err != nil {
		return err
	}

	description.field = description.field.WithDefault(expression)
	return nil
}
