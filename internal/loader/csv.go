package loader

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/table"
)

func NewTableFromCSV(name, filepath string, fields []shared.Field) (*table.Table, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	rawValues, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, err
	}

	convertedValues, err := convertRecords(rawValues, fields)
	if err != nil {
		return nil, err
	}

	var tableFields []shared.Field
	for _, field := range fields {
		tableFields = append(tableFields, field.WithRelationName(name))
	}

	rows, err := shared.NewRowsWithValues(tableFields, convertedValues)
	if err != nil {
		return nil, err
	}

	table := table.NewTable(name, rows.Fields)

	for i := 0; i < rows.Size(); i++ {
		if _, err := table.Insert(rows.Row(i)); err != nil {
			return nil, err
		}
	}

	return table, nil
}

func convertRecords(rawValues [][]string, fields []shared.Field) ([][]any, error) {
	convertedValues := make([][]any, 0, len(rawValues))
	for _, rawValues := range rawValues {
		if len(rawValues) != len(fields) {
			return nil, fmt.Errorf("expected %d values, got %d", len(fields), len(rawValues))
		}

		values, err := convertValues(rawValues, fields)
		if err != nil {
			return nil, err
		}

		convertedValues = append(convertedValues, values)
	}

	return convertedValues, nil
}

func convertValues(rawValues []string, fields []shared.Field) ([]any, error) {
	values := make([]any, 0, len(rawValues))
	for i, rawValue := range rawValues {
		value, err := convertValue(rawValue, fields[i].Type())
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

func convertValue(rawValue string, typ shared.Type) (any, error) {
	if rawValue == "NULL" {
		if !typ.Nullable {
			return nil, fmt.Errorf("expected non-nullable value, got NULL")
		}

		return nil, nil
	}

	switch typ.Kind {
	case shared.TypeKindText:
		return rawValue, nil

	case shared.TypeKindNumeric:
		return strconv.Atoi(rawValue)

	case shared.TypeKindBool:
		return rawValue == "true", nil
	}

	return nil, fmt.Errorf("unconvertible type %s", typ.Kind)
}
