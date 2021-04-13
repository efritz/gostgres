package relations

type limitRelation struct {
	Relation
	limit int
}

var _ Relation = &limitRelation{}

func NewLimit(table Relation, limit int) Relation {
	return &limitRelation{
		Relation: table,
		limit:    limit,
	}
}

func (r *limitRelation) Scan(scanContext ScanContext, visitor VisitorFunc) error {
	invocations := 0
	return r.Relation.Scan(scanContext, func(scanContext ScanContext, values []interface{}) (bool, error) {
		if invocations >= r.limit {
			return false, nil
		}

		invocations++
		return visitor(scanContext, values)
	})
}
