package scan

import (
	"github.com/efritz/gostgres/internal/shared/rows"
)

func ScanIntoRows(scanner Scanner, target rows.Rows) (rows.Rows, error) {
	if err := VisitRows(scanner, func(row rows.Row) (bool, error) {
		var err error
		target, err = target.AddValues(row.Values)
		return true, err
	}); err != nil {
		return rows.Rows{}, err
	}

	return target, nil
}

type VisitorFunc func(row rows.Row) (bool, error)

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
