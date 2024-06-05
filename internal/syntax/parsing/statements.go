package parsing

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/explain"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

func (p *parser) initDDLParsers() {
	p.ddlParsers = ddlParsers{
		tokens.TokenTypeCreate: p.parseCreate,
		tokens.TokenTypeAlter:  p.parseAlter,
	}
}

func (p *parser) initStatementParsers() {
	p.explainableParsers = explainableParsers{
		tokens.TokenTypeSelect: p.parseSelect,
		tokens.TokenTypeInsert: p.parseInsert,
		tokens.TokenTypeUpdate: p.parseUpdate,
		tokens.TokenTypeDelete: p.parseDelete,
	}
}

// statement := ddlStatement | ( [ `EXPLAIN` ] explainableStatement )
// ddlStatement := ( `CREATE` createTail ) | ( `ALTER` alterTail )
// explainableStatement := [ `WITH` name `AS` `(` selectInsertUpdateOrDelete `)` [, ...] ] selectInsertUpdateOrDelete
func (p *parser) parseStatement() (Query, error) {
	for tokenType, parser := range p.ddlParsers {
		token := p.current()
		if p.advanceIf(isType(tokenType)) {
			return parser(token)
		}
	}

	isExplain := p.advanceIf(isType(tokens.TokenTypeExplain))

	type namedNode struct {
		name string
		node queries.Node
	}
	var namedNodes []namedNode
	if p.advanceIf(isType(tokens.TokenTypeWith)) {
		for {
			name, err := p.parseIdent()
			if err != nil {
				return nil, err
			}

			if _, err := p.mustAdvance(isType(tokens.TokenTypeAs)); err != nil {
				return nil, err
			}

			node, err := parseParenthesized(p, func() (queries.Node, error) {
				return p.parseSelectInsertUpdateOrDelete()
			})
			if err != nil {
				return nil, err
			}

			namedNodes = append(namedNodes, namedNode{name, node})

			if !p.advanceIf(isType(tokens.TokenTypeComma)) {
				break
			}
		}
	}

	node, err := p.parseSelectInsertUpdateOrDelete()
	if err != nil {
		return nil, err
	}

	if len(namedNodes) > 0 {
		fmt.Printf("> %#v\n", namedNodes)
	}

	if isExplain {
		node = explain.NewExplain(node)
	}

	return queries.NewQuery(node), nil
}

// selectInsertUpdateOrDelete := ( `SELECT` selectTail ) | ( `INSERT` insertTail ) | ( `UPDATE` updateTail ) | ( `DELETE` deleteTail )
func (p *parser) parseSelectInsertUpdateOrDelete() (queries.Node, error) {
	for tokenType, parser := range p.explainableParsers {
		token := p.current()
		if p.advanceIf(isType(tokenType)) {
			return parser(token)
		}
	}

	return nil, fmt.Errorf("expected start of select, insert, update, or delete statement (near %s)", p.current().Text)
}
