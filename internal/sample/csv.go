package sample

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/table"
)

func populateTableFromCSV(table *table.Table, filepath string) error {
	var tableFields []shared.Field
	for _, field := range table.Fields() {
		if field.Internal() {
			continue
		}

		tableFields = append(tableFields, field.WithRelationName(table.Name()))
	}

	file, err := os.Open(filepath)
	if err != nil {
		return err
	}

	rawValues, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return err
	}

	convertedValues, err := convertRecords(rawValues, tableFields)
	if err != nil {
		return err
	}

	rows, err := shared.NewRowsWithValues(tableFields, convertedValues)
	if err != nil {
		return err
	}

	for i := 0; i < rows.Size(); i++ {
		if _, err := table.Insert(rows.Row(i)); err != nil {
			return err
		}
	}

	return nil
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

	case shared.TypeKindTimestampTz:
		return time.Parse("2006-01-02 15:04:05-07", rawValue)
	}

	return nil, fmt.Errorf("unconvertible type %s", typ.Kind)
}
