package mongo

import (
	"time"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"github.com/matoous/go-nanoid/v2"
)

const alphabet = "0123456789abcdefghijclmnopqrstuvwxyz"

const maxLen = 10

// ensureId adds the id property when missing or when it's an empty string.
func ensureId(data bson.M, idPrefix string) string {
	maybeId, hasId := data["_id"]

	var id, finalId string
	if !hasId || maybeId == "" {
		id, _ = gonanoid.Generate(alphabet, maxLen)
		if idPrefix != "" {
			finalId = fmt.Sprintf("%s_%s", idPrefix, id)
		} else {
			finalId = id
		}

		data["_id"] = finalId
	}

	return finalId
}

func ensureNoId(data bson.M) {
	if _, ok := data["_id"]; ok {
		delete(data, "_id")
	}
}

func ensureExistingCreatedAt(data bson.M, old bson.M) time.Time {
	var created time.Time
	createdAny, ok := old["createdAt"]
	if ok {
		switch val := createdAny.(type) {
		case bson.DateTime:
			created = val.Time().UTC()
		default:
			created = createdAny.(time.Time)
		}
		
		data["createdAt"] = created
	}

	return created
}

func ensureCreatedAt(data bson.M) time.Time {
	now := time.Now().UTC().Truncate(time.Millisecond)

	data["createdAt"] = now

	return now
}

func ensureUpdatedAt(data bson.M) time.Time {
	now := time.Now().UTC().Truncate(time.Millisecond)

	data["updatedAt"] = now
	
	return now
}