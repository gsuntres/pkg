package mongo

import (
	"context"
		
	"go.mongodb.org/mongo-driver/v2/bson"
	"github.com/gsuntres/pkg/commons"
)

type FindOptions struct {
	Offset int64   `json:"offset"`
	Limit  int64   `json:"limit"`
	Alias  string  `json:"alias"`
	Sort   [][]any `json:"sort"`
}

func FindOptionsFromMap(m map[string]any) *FindOptions {
	opts := &FindOptions{}

	limit, hasLimit := commons.GetGeneric[int64](m, "limit") 
	if hasLimit {
		if limit == 0 {
			opts.Limit = 10
		} else {
			opts.Limit = limit
		}
	}

	offset, hasOffset := commons.GetGeneric[int64](m, "offset") 
	if hasOffset {
		opts.Offset = offset
	}

	alias := commons.GetString(m, "alias")
	opts.Alias = alias
	
	sort, hasSort := commons.GetGeneric[[][]any](m, "sort")
	if hasSort {
		opts.Sort = sort
	}

	return opts
}

// Find is used to fetch the first page of data.
func (c *MongoClient) Find(ctx context.Context, database, name string, filter bson.M, opts *FindOptions) (bson.M, error) {

	var limit int64

	if opts != nil {
		limit = opts.Limit
	}

	// 1. Prepare to query.
	finalName := name
	if opts != nil && commons.StringIsNotBlank(opts.Alias) {
		finalName = opts.Alias
	}
	collection := c.GetCollection(database, finalName)

	pageSize := max(limit, c.Limit)

	sortOpts := commons.BsonDFromSlice(opts.Sort)
	var sort bson.D

	if len(sortOpts) == 0 {
		sort = sortOpts
	} else {
		sort = bson.D{{Key: "_id", Value: 1}}
	}
	
	result, err := c.DiscriminatorCheckAndApplyToFilter(ctx, name, filter)
	if err != nil {
		return nil, err
	}

	f := Mongofy(&Query{ 
		Filter: filter,
		WithDiscriminator: result.WithDiscriminator,
		DiscriminatorField: result.DiscriminatorField,	
	})

	pipeline := BuildPaginationPipeline(0, pageSize + 1, f, sort)

	// 2. Query
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 3. Build results
	var facetResults []bson.M
	if err = cursor.All(ctx, &facetResults); err != nil {
		return nil, err
	}

	root := facetResults[0]

	data := root["data"].(bson.A)

	metadata := root["metadata"].(bson.A)

	var totalValue any
	if len(metadata) != 0 {
		switch metadataRoot := metadata[0].(type) {
		case bson.D: 
			totalValue, _ = commons.BsonDGetAny(metadataRoot, "total")
		case bson.M:
			totalValue = metadataRoot["total"]
		}
	}

	var total int64
	switch v := totalValue.(type) {
	case int32:
	  total = int64(v)
	case int64:
	  total = v
	default:
		total = 0
	}

	hasMore := false
	if int64(len(data)) > pageSize {
		hasMore = true
		data = data[:pageSize]
	}

	out := bson.M{
		"data": data,
		"has_more": hasMore,
		"total": total,
	}

	if hasMore {
		// next cursor
		var last bson.M = data[len(data) - 1].(bson.M)
		var nextCursor string
		lastId := last["_id"]

		switch lid := lastId.(type) {
		case bson.ObjectID:
			nextCursor = EncodeCursor(lid)
		case string:
			nextCursor = lid
		}

		out["next_cursor"] = nextCursor
	}

	return out, nil
}

func (c *MongoClient) FindOffset(ctx context.Context, database, name string, filter bson.M, opts *FindOptions) (bson.M, error) {
	var offset int64
	var limit int64

	if opts != nil {
		limit = opts.Limit
	}

	if opts != nil {
		offset = opts.Offset
	}

	// 1. Prepare to query.
	finalName := name
	if opts != nil && commons.StringIsNotBlank(opts.Alias) {
		finalName = opts.Alias
	}
	collection := c.GetCollection(database, finalName)

	finalLimit := max(limit, c.Limit)

	result, err := c.DiscriminatorCheckAndApplyToFilter(ctx, name, filter)
	if err != nil {
		return nil, err
	}

	f := Mongofy(&Query{ 
		Filter: filter,
		WithDiscriminator: result.WithDiscriminator,
		DiscriminatorField: result.DiscriminatorField,	
	})

	sort := commons.BsonDFromSlice(opts.Sort)
	
	pipeline := BuildPaginationPipeline(offset, finalLimit, f, sort)

	// 2. Query
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 3. Build results
	var facetResults []bson.M
	if err = cursor.All(ctx, &facetResults); err != nil {
		return nil, err
	}

	root := facetResults[0]

	data := root["data"].(bson.A)

	metadata := root["metadata"].(bson.A)

	var totalValue any
	if len(metadata) != 0 {
		switch metadataRoot := metadata[0].(type) {
		case bson.D: 
			totalValue, _ = commons.BsonDGetAny(metadataRoot, "total")
		case bson.M:
			totalValue = metadataRoot["total"]
		}
	}

	var total int64
	switch v := totalValue.(type) {
	case int32:
	  total = int64(v)
	case int64:
	  total = v
	default:
		total = 0
	}

	hasMore := false
	if total > offset + finalLimit {
		hasMore = true
	}

	hasPrevious := false
	if finalLimit - offset < 0 {
		hasPrevious = true
	}

	out := bson.M{
		"data": data,
		"offset": offset,
		"limit": finalLimit,
		"has_more": hasMore,
		"has_previous": hasPrevious,
		"total": total,
	}

	return out, nil
}
