package scan

import (
	"github.com/efritz/gostgres/internal/shared"
)

func ScanIntoRows(scanner Scanner, rows shared.Rows) (shared.Rows, error) {
	if err := VisitRows(scanner, func(row shared.Row) (bool, error) {
		var err error
		rows, err = rows.AddValues(row.Values)
		return true, err
	}); err != nil {
		return shared.Rows{}, err
	}

	return rows, nil
}

type VisitorFunc func(row shared.Row) (bool, error)

func VisitRows(scanner Scanner, visitor VisitorFunc) error {
	for {
		row, err := scanner.Scan()
		if err != nil {
			if err == ErrNoRows {
				break
			}

			return err
		}

		if ok, err := visitor(row); err != nil {
			return err
		} else if !ok {
			break
		}
	}

	return nil
}
