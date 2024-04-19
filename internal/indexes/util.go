package indexes

import "github.com/efritz/gostgres/internal/shared"

// TODO - deduplicate
func extractTID(row shared.Row) (int, bool) {
	if len(row.Fields) == 0 || row.Fields[0].Name != "tid" {
		return 0, false
	}

	tid, ok := row.Values[0].(int)
	return tid, ok
}
