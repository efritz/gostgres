package parsing

import (
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/queries/ddl"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// createSequenceTail := ident [ `AS` basicType ]
func (p *parser) parseCreateSequence() (queries.Query, error) {
	name, err := p.parseIdent()
	if err != nil {
		return nil, err
	}

	typ := shared.TypeBigInteger
	if p.advanceIf(isType(tokens.TokenTypeAs)) {
		typ, err = p.parseBasicType()
		if err != nil {
			return nil, err
		}
	}

	// TODO - increment, min/max value, start, cycle
	return ddl.NewCreateSequence(name, typ), nil
}
