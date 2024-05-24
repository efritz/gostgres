package parsing

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/ddl"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// alterTableTail := ident `ADD CONSTRAINT` addConstraintTail
func (p *parser) parseAlterTable() (queries.Query, error) {
	tableName, err := p.parseIdent()
	if err != nil {
		return nil, err
	}

	if p.advanceIf(isType(tokens.TokenTypeAdd), isType(tokens.TokenTypeConstraint)) {
		return p.parseAddConstraint(tableName)
	}

	return nil, fmt.Errorf("unexpected alter table statement (near %s)", p.current().Text)

}

func (p *parser) initConstraintParsers() {
	p.addConstraintParsers = addConstraintParsers{
		tokens.TokenTypePrimaryKey: p.parsePrimaryKeyConstraint,
		tokens.TokenTypeForeignKey: p.parseForeignKeyConstraint,
		tokens.TokenTypeCheck:      p.parseCheckConstraint,
	}
}

// addConstraintTail := ident ( ( `PRIMARY KEY` primaryKeyConstraintTail ) | ( `FOREIGN KEY` foreignKeyConstraintTail ) | ( `CHECK` checkConstraintTail ) )
func (p *parser) parseAddConstraint(tableName string) (queries.Query, error) {
	name, err := p.parseIdent()
	if err != nil {
		return nil, err
	}

	for tokenType, parser := range p.addConstraintParsers {
		if p.advanceIf(isType(tokenType)) {
			return parser(name, tableName)
		}
	}

	return nil, fmt.Errorf("expected constraint definition (near %s)", p.current().Text)
}

// primaryKeyConstraintTail := `(` ident [, ...] `)`
func (p *parser) parsePrimaryKeyConstraint(name, tableName string) (queries.Query, error) {
	columnNames, err := parseParenthesizedCommaSeparatedList(p, false, false, p.parseIdent)
	if err != nil {
		return nil, err
	}

	// TODO - also ensure column names are not-null
	return ddl.NewCreatePrimaryKeyConstraint(name, tableName, columnNames), nil
}

// foreignKeyConstraintTail := `(` ident [, ...] `)` `REFERENCES` ident `(` ident [, ...] `)`
func (p *parser) parseForeignKeyConstraint(name, tableName string) (queries.Query, error) {
	columnNames, err := parseParenthesizedCommaSeparatedList(p, false, false, p.parseIdent)
	if err != nil {
		return nil, err
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeReferences)); err != nil {
		return nil, err
	}

	refTable, err := p.parseIdent()
	if err != nil {
		return nil, err
	}

	refColumnNames, err := parseParenthesizedCommaSeparatedList(p, false, false, p.parseIdent)
	if err != nil {
		return nil, err
	}

	return ddl.NewCreateForeignKeyConstraint(name, tableName, columnNames, refTable, refColumnNames), nil
}

// checkConstraintTail := `(` expression `)`
func (p *parser) parseCheckConstraint(name, tableName string) (queries.Query, error) {
	expr, err := parseParenthesized(p, p.parseRootExpression)
	if err != nil {
		return nil, err
	}

	return ddl.NewCreateCheckConstraint(name, tableName, expr), nil
}
