package parsing

import (
	"fmt"
	"strings"

	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/queries/ddl"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// create := `TABLE` name `(` [ column [, ...] ] `)`
func (p *parser) parseCreate(token tokens.Token) (queries.Node, error) {
	if _, err := p.mustAdvance(isType(tokens.TokenTypeTable)); err != nil {
		return nil, err
	}

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
		typ = shared.TypeText
	case "integer":
		typ = shared.TypeNumeric
	case "boolean":
		typ = shared.TypeBool
	default:
		return shared.Field{}, fmt.Errorf("unknown type %s", dataType.Text)
	}

	if p.advanceIf(isType(tokens.TokenTypeNotNull)) {
		typ = typ.NonNullable()
	}

	return shared.NewField("", name.Text, typ), nil
}
