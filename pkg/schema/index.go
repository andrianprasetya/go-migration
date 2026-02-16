package schema

// IndexDefinition represents an index on a database table.
type IndexDefinition struct {
	Name    string
	Columns []string
	Unique  bool
}
