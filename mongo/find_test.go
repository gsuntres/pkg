package mongo

import (
	"os"
	"context"
	"testing"
	"encoding/json"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gsuntres/pkg/commons"
)

func TestFind_Default(t *testing.T) {
	client := GetMongoClient()

	filter := bson.M{"name": bson.M{"$regex": "OSRAM"}}
	findResult, err := client.Find(context.Background(), "mydb", "mycollection", filter, &FindOptions{
		Offset: int64(0),
	})
	if err != nil {
		t.Fatalf("Failed to insertOne %#v", err)
	}

	dataAny, hasData := findResult["data"]
	if !hasData {
		t.Fatal("no data")
	}

	data := dataAny.(bson.A)

	if len(data) != 1 {
		t.Fatalf("Expected to return 1 document but got %d", len(data))
	}

	hasMoreAny, hasMoreOk := findResult["has_more"]
	if !hasMoreOk {
		t.Fatal("no has more")
	}

	hasMore := hasMoreAny.(bool)

	if hasMore {
		t.Fatalf("Expected to have reached the end of the results")
	}

	totalAny, totalOk := findResult["total"]
	if !totalOk {
		t.Fatal("no total")
	}

	total := totalAny.(int64)

	if total != 1 {
		t.Fatalf("Expected total to be 1 but found %d", total)
	}
}

func TestFind_Discriminator(t *testing.T) {
	// 1. Register schemas
	schemaStore, err := os.ReadFile("./.test/store.json")
	if err != nil { t.Fatal(err) }

	schemaOffer, err := os.ReadFile("./.test/offer.json")
	if err != nil { t.Fatal(err) }

	var store bson.M
	if err := json.Unmarshal(schemaStore, &store); err != nil {
		t.Fatalf("Length: %d, First bytes: %x\n", len(schemaStore), schemaStore[:4])
	}

	var offer bson.M
	if err := json.Unmarshal(schemaOffer, &offer); err != nil {
		t.Fatalf("Length: %d, First bytes: %x\n", len(schemaOffer), schemaOffer[:4])
	}

	client := GetMongoClient()
	client.AddDefinition(store)
	client.AddDefinition(offer)

	// Store str_1234 has OSRAM 1.
	ctx1 := context.Background()
	ctx1 = context.WithValue(ctx1, "store", "str_1234")
  offer1 := map[string]any { "name": "OSRAM 1" }
  _, err = client.InsertOne(ctx1, "mydb", "offer", offer1)
  if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

  // Store str_4321 has OSRAM 2
  ctx2 := context.Background()
	ctx2 = context.WithValue(ctx2, "store", "str_4321")
	offer2 := map[string]any { "name": "OSRAM 2" }
  _, err = client.InsertOne(ctx2, "mydb", "offer", offer2)
  if err != nil { t.Fatalf("Failed to insertOne %#v", err) }
	
	// Within store str_1234 searching for OSRAM should return one result
	filter := bson.M{"name": bson.M{"$regex": "OSRAM*"}}
	findResult, err := client.Find(ctx1, "mydb", "offer", filter, &FindOptions{	Offset: int64(0) })
	if err != nil { t.Fatalf("Failed to find %#v", err) }

	dataAny, hasData := findResult["data"]
	if !hasData { t.Fatal("no data") }

	data := dataAny.(bson.A)

	if len(data) != 1 {
		t.Fatalf("Expected to return 1 document but got %d", len(data))
	}

	arr, _ := commons.BsonAToSlice(data)

	o := arr[0]

	name := o["name"]

	if name != "OSRAM 1" {
	 	t.Fatalf("Expected OSRAM 1 not %s", name)
	}
}
