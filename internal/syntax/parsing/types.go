package parsing

import (
	"fmt"
	"strings"

	"github.com/efritz/gostgres/internal/shared"
)

// basicType := ident
func (p *parser) parseBasicType() (shared.Type, error) {
	dataType, err := p.parseIdent()
	if err != nil {
		return shared.Type{}, err
	}

	var typ shared.Type
	switch strings.ToLower(dataType) {
	case "text":
		typ = shared.TypeText
	case "smallint":
		typ = shared.TypeSmallInteger
	case "integer":
		typ = shared.TypeInteger
	case "bigint":
		typ = shared.TypeBigInteger
	case "real":
		typ = shared.TypeReal
		// TODO - use multi-phrase keyword
	case "double":
		if !p.advanceIf(isIdent("precision")) {
			return shared.Type{}, fmt.Errorf("unknown type %q", "double")
		}
		typ = shared.TypeDoublePrecision
	case "numeric":
		typ = shared.TypeNumeric
	case "boolean":
		typ = shared.TypeBool
		// TODO - use multi-phrase keyword(s)
	case "timestamp":
		if !p.advanceIf(isIdent("with"), isIdent("time"), isIdent("zone")) {
			return shared.Type{}, fmt.Errorf("unknown type %q", "timestamp")
		}
		typ = shared.TypeTimestampTz
	default:
		return shared.Type{}, fmt.Errorf("unknown type %s", dataType)
	}

	return typ, nil
}
