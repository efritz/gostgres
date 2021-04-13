package shared

type Row struct {
	Fields []Field
	Values []interface{}
}

type Rows struct {
	Fields []Field
	Values [][]interface{}
}

type Field struct {
	Name         string
	RelationName string
	// TODO - value types
}
