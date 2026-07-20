package mongo

// TODO(me): Place it in find.go?
type FindRequest struct {
	Entity string
	Limit int
	Skip int
	Filter map[string]any
}

// TODO(me): Remove it in favor of [FindResult]?
type DataResults struct {
	Data []map[string]any
}

// TODO(me): Remove it in favor of [FindResult]?
type Metadata struct {
  Total int64 `bson:"total"`
}
// TODO(me): Remove it in favor of [FindResult]?
type PaginatedResponse struct {
  Metadata []Metadata `bson:"metadata" json:"metadata"`
  // Initializing this as an empty slice literal []User{}
  // prevents "null" in your JSON output.
  Data []map[string]any `bson:"data" json:"data"`
}
