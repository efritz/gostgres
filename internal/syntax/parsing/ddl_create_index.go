package parsing

import (
	"github.com/efritz/gostgres/internal/execution/queries/ddl"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// createIndexTail := ident `ON` ident createIndexUsing `(` ( expressionWithDirection [, ...] ) `)` where
func (p *parser) parseCreateIndex(unique bool) (Query, error) {
	name, err := p.parseIdent()
	if err != nil {
		return nil, err
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeOn)); err != nil {
		return nil, err
	}

	tableName, err := p.parseIdent()
	if err != nil {
		return nil, err
	}

	method, err := p.parseCreateIndexUsing()
	if err != nil {
		return nil, err
	}

	columnExpressions, err := parseParenthesizedCommaSeparatedList(p, false, false, p.parseExpressionWithDirection)
	if err != nil {
		return nil, err
	}

	where, _, err := p.parseWhere()
	if err != nil {
		return nil, err
	}

	// TODO - concurrently
	// TODO - if not exists
	// TODO - include
	// TODO - NULLS FIRST | LAST
	// TODO - nulls distinct
	return ddl.NewCreateIndex(name, tableName, method, unique, columnExpressions, where), nil
}

// createIndexUsing := `USING` ( `btree` | `hash` )
func (p *parser) parseCreateIndexUsing() (string, error) {
	if p.advanceIf(isType(tokens.TokenTypeUsing)) {
		return p.parseIdent()
	}

	return "btree", nil
}
