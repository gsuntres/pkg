package mongo

import (
  "go.mongodb.org/mongo-driver/v2/bson"
  "go.mongodb.org/mongo-driver/v2/mongo"
)

func BuildPaginationPipeline(skip, limit int64, filter bson.M, sort bson.D) mongo.Pipeline {
  pipe := mongo.Pipeline{
    // 1. GLOBAL FILTER: Always filter first to use indexes
    {{Key: "$match", Value: filter}},
  }

  // 2. GLOBAL SORT: Sort here so both 'total' and 'data' facets use the same order
  if sort != nil && len(sort) > 0 {
    pipe = append(pipe, bson.D{{Key: "$sort", Value: sort}})
  }

  // 3. FACET: Split the pipeline into two parallel paths
  pipe = append(pipe, 
    bson.D{{Key: "$facet", Value: bson.D{
      // Path A: Get the total count of documents matching the filter
      {Key: "metadata", Value: mongo.Pipeline{
          {{Key: "$count", Value: "total"}},
      }},
      // Path B: Get the specific page of data
      {Key: "data", Value: mongo.Pipeline{
          {{Key: "$skip", Value: skip}},
          {{Key: "$limit", Value: limit}},
      }},
    }}},
  )

  return pipe
}

func BuildPaginationPipelineNext(limit int64, filter bson.M, sort bson.M) mongo.Pipeline {
  
  return mongo.Pipeline{
    // 1. GLOBAL FILTER: Always filter first to use indexes
    {{Key: "$match", Value: filter}},
    
    // 2. GLOBAL SORT: Sort here so both 'total' and 'data' facets use the same order
    {{Key: "$sort", Value: sort}},

    // 3. FACET: Split the pipeline into two parallel paths
    {{Key: "$facet", Value: bson.D{
      // Path A: Get the total count of documents matching the filter
      {Key: "metadata", Value: mongo.Pipeline{
          {{Key: "$count", Value: "total"}},
      }},
      // Path B: Get the specific page of data
      {Key: "data", Value: mongo.Pipeline{
          {{Key: "$limit", Value: limit}},
      }},
    }}},
  }
}