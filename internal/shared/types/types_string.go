package types

import "time"

const (
	dateFormat      = "2006-01-02"
	timestampFormat = "2006-01-02 15:04:05-07"
)

func refineString(value string, typ Type) (Type, any, bool) {
	if typ == TypeTimestampTz {
		if t, err := time.Parse(dateFormat, value); err == nil {
			return typ, t, true // Just date parsing
		} else if t, err := time.Parse(timestampFormat, value); err == nil {
			return typ, t, true
		} else {
			return TypeUnknown, nil, false
		}
	}

	return typ, value, typ == TypeAny || typ == TypeText
}
