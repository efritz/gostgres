package loader

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/table"
)

func NewTableFromCSV(name, filepath string, fieldDescriptions []FieldDescription) (*table.Table, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	rawValues, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, err
	}

	convertedValues, err := convertRecords(rawValues, fieldDescriptions)
	if err != nil {
		return nil, err
	}

	rows, err := shared.NewRowsWithValues(convertFields(name, fieldDescriptions), convertedValues)
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

func convertRecords(rawValues [][]string, fieldDescriptions []FieldDescription) ([][]any, error) {
	convertedValues := make([][]any, 0, len(rawValues))
	for _, rawValues := range rawValues {
		if len(rawValues) != len(fieldDescriptions) {
			return nil, fmt.Errorf("expected %d values, got %d", len(fieldDescriptions), len(rawValues))
		}

		values, err := convertValues(rawValues, fieldDescriptions)
		if err != nil {
			return nil, err
		}

		convertedValues = append(convertedValues, values)
	}

	return convertedValues, nil
}

func convertValues(rawValues []string, fieldDescriptions []FieldDescription) ([]any, error) {
	values := make([]any, 0, len(rawValues))
	for i, rawValue := range rawValues {
		value, err := convertValue(rawValue, fieldDescriptions[i].TypeKind)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

func convertValue(rawValue string, typeKind shared.TypeKind) (any, error) {
	switch typeKind {
	case shared.TypeKindText:
		return rawValue, nil

	case shared.TypeKindNumeric:
		return strconv.Atoi(rawValue)

	case shared.TypeKindBool:
		return rawValue == "true", nil
	}

	return nil, fmt.Errorf("unconvertible type %s", typeKind)
}

func convertFields(name string, fieldDescriptions []FieldDescription) []shared.Field {
	fields := make([]shared.Field, 0, len(fieldDescriptions))
	for _, fieldDescription := range fieldDescriptions {
		fields = append(fields, shared.NewField(name, fieldDescription.Name, fieldDescription.TypeKind, false))
	}

	return fields

}
