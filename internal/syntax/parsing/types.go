package parsing

import (
	"fmt"
	"strings"

	"github.com/efritz/gostgres/internal/shared/types"
)

// basicType := ident
func (p *parser) parseBasicType() (types.Type, error) {
	dataType, err := p.parseIdent()
	if err != nil {
		return types.TypeUnknown, err
	}

	var typ types.Type
	switch strings.ToLower(dataType) {
	case "text":
		typ = types.TypeText
	case "smallint":
		typ = types.TypeSmallInteger
	case "integer":
		typ = types.TypeInteger
	case "bigint":
		typ = types.TypeBigInteger
	case "real":
		typ = types.TypeReal
		// TODO - use multi-phrase keyword
	case "double":
		if !p.advanceIf(isIdent("precision")) {
			return types.TypeUnknown, fmt.Errorf("unknown type %q", "double")
		}
		typ = types.TypeDoublePrecision
	case "numeric":
		typ = types.TypeNumeric
	case "boolean":
		typ = types.TypeBool
		// TODO - use multi-phrase keyword(s)
	case "timestamp":
		if !p.advanceIf(isIdent("with"), isIdent("time"), isIdent("zone")) {
			return types.TypeUnknown, fmt.Errorf("unknown type %q", "timestamp")
		}
		typ = types.TypeTimestampTz
	default:
		return types.TypeUnknown, fmt.Errorf("unknown type %s", dataType)
	}

	return typ, nil
}
