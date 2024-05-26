package parsing

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

func (p *parser) initCreateParsers() {
	p.createParsers = createParsers{
		tokens.TokenTypeTable:    p.parseCreateTable,
		tokens.TokenTypeSequence: p.parseCreateSequence,
		tokens.TokenTypeIndex:    func() (queries.Query, error) { return p.parseCreateIndex(false) },
	}
}

// createTail := ( `TABLE` createTableTail ) | ( `SEQUENCE` createSequenceTail ) | ( [ `UNIQUE` ] `INDEX` createIndexTail )
func (p *parser) parseCreate(token tokens.Token) (queries.Query, error) {
	for tokenType, parser := range p.createParsers {
		if p.advanceIf(isType(tokenType)) {
			return parser()
		}
	}

	// TODO - make unique index one token?
	if p.advanceIf(isType(tokens.TokenTypeUnique), isType(tokens.TokenTypeIndex)) {
		return p.parseCreateIndex(true)
	}

	return nil, fmt.Errorf("expected create statement (near %s)", p.current().Text)
}

func (p *parser) initAlterParsers() {
	p.alterParsers = alterParsers{
		tokens.TokenTypeTable: p.parseAlterTable,
	}
}

// alterTail := `TABLE` alterTableTail
func (p *parser) parseAlter(token tokens.Token) (queries.Query, error) {
	for tokenType, parser := range p.alterParsers {
		if p.advanceIf(isType(tokenType)) {
			return parser()
		}
	}

	return nil, fmt.Errorf("expected alter statement (near %s)", p.current().Text)
}
