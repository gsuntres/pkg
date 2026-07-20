package mongo

import (
	"os"
	"time"
	"reflect"
	"testing"
  "context"
  "fmt"
	"encoding/json"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/bson"

  "github.com/testcontainers/testcontainers-go"
  "github.com/testcontainers/testcontainers-go/modules/mongodb"

  "github.com/gsuntres/pkg/commons"
)

type Sample struct {
	Id   			string `json:"id" bson:"_id,omitempty"`
	Name 		  string `json:"name" bson:"name"`
	Age       int32 `json:"age"  bson:"age"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

func TestMain(m *testing.M) {
  if os.Getenv("RUN_INTEGRATION") == "" {
    fmt.Println("Skipping package tests: RUN_INTEGRATION is missing")
    os.Exit(0) // Exit with success, but no tests ran
  }

  ctx := context.Background()

  user := "admin"
  pass := "1"

  // 1. Setup: Start the MongoDB container
  mongoContainer, err := mongodb.Run(ctx, 
  	"mongo:8",
  	mongodb.WithReplicaSet("rs0"),
    testcontainers.WithEnv(map[string]string{
      "MONGO_INITDB_ROOT_USERNAME": user,
      "MONGO_INITDB_ROOT_PASSWORD": pass,      
    }),
  )
  if err != nil {
    panic("failed to start container")
  }

  // 2. Get the connection string dynamically
  endpoint, _ := mongoContainer.ConnectionString(ctx)

	mongoDebug := os.Getenv("DEBUG") != ""

  Start(&MongoStartProps{
    MongoUri: endpoint,
    MongoUser: user,
    MongoPass: pass,
    ContextFields: []string{"account", "store"},
    MongoDebugQuery: mongoDebug,
    WithAudit: true,
  })

  // 3. Run tests
  LoadTestSample()

  code := m.Run()

  // 4. Teardown: Clean up resources
  Stop()

  _ = testcontainers.TerminateContainer(mongoContainer)

  os.Exit(code)
}

func TestApplySchema(t *testing.T) {
	// prepare schema sample
	schemaStr := `
	{
		"bsonType": "object",
		"properties": {
			"name": {
				"bsonType": "string"
			},
			"active": {
				"bsonType": "boolean"
			}
		},
		"required": ["name"],
		"additionalProperties": true
	}
	`

	var result bson.M
	err := json.Unmarshal([]byte(schemaStr), &result)
	if err != nil {
		t.Fatalf("%v", err)
	}

	cd := &CollectionDefinition{
		Schema: result,
	}

	opts := options.CreateCollection()

	ApplySchema(cd, opts)

	options := &options.CreateCollectionOptions{}

	for _, o := range opts.List() {
		o(options)
	}

	expected := bson.M{"$jsonSchema":bson.M{"additionalProperties":true, "bsonType":"object", "properties":bson.D{bson.E{Key:"active", Value:bson.D{bson.E{Key:"bsonType", Value:"boolean"}}}, bson.E{Key:"name", Value:bson.D{bson.E{Key:"bsonType", Value:"string"}}}}, "required":bson.A{"name"}}}

	aj, _ := json.Marshal(options.Validator)
	ej, _ := json.Marshal(expected)

	var am, em any
	json.Unmarshal(aj, &am)
	json.Unmarshal(ej, &em)

	if !reflect.DeepEqual(am, em) {
		t.Fatalf("Validator should have been set")
	}
}

func TestGetCollection_Default(t *testing.T) {
	c := GetMongoClient()

	col := c.GetCollection("mydb", "mycol")

	if col == nil {
		t.Fatal("should have return collection")
	}

	if col.Name() != "mycol" {
		t.Fatal("Wrong name")
	}
}

func TestDisableValidation(t *testing.T) {
	// 1. Register schemas
	schema, err := os.ReadFile("./.test/singleton.json")
	if err != nil { t.Fatal(err) }

	var singleton bson.M
	if err := json.Unmarshal(schema, &singleton); err != nil {
		// t.Fatalf("Length: %d, First bytes: %x\n", len(schema), schema[:4])
		_, help := commons.JsonSprintError(err, schema)
		t.Fatalf("Failed to unmarshal %s", help)
	}

	client := GetMongoClient()
	client.AddDefinition(singleton)

	// 2. Insert data
	p1 := map[string]any {
		"age": int64(67),
	}

	o, err := client.UpdateInc(context.Background(), 
		"mydb", 
		"singleton", 
		map[string]any {"type": "counter","name": "MyName112"}, 
		p1,
	)
	if err != nil { t.Fatalf("Failed to UpdateInc %#v", err) }

	age, _ := commons.GetGeneric[int64](o, "age")
	if age != 67 {
		t.Fatalf("%d", age)
	}	
}
