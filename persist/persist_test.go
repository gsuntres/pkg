package persist

import (
	"os"
	"testing"
  "context"
  "fmt"

  "github.com/testcontainers/testcontainers-go"
  "github.com/testcontainers/testcontainers-go/modules/mongodb"
  "go.mongodb.org/mongo-driver/v2/bson"

  "github.com/gsuntres/pkg/mongo"
  "github.com/gsuntres/pkg/structful"

)

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
    }),)
  if err != nil {
    panic("failed to start container")
  }

  // 2. Get the connection string dynamically
  endpoint, _ := mongoContainer.ConnectionString(ctx)

  mongo.Start(&mongo.MongoStartProps{
    MongoUri: endpoint,
    MongoUser: user,
    MongoPass: pass,
  })

  // Initialize persist
  structful.Init(&structful.InitProps{
    MongoSysDb: "db_sys",
    Version: "0",
    MongoStructfulCollection: "_structful",
  })

  structful.SeedStructful([]string{
    "./.test/structful/item.json",
  })

  Init(&InitProps{})

  // 3. Run tests
  code := m.Run()

  // 4. Teardown: Clean up resources
  mongo.Stop()
  _ = testcontainers.TerminateContainer(mongoContainer)

  os.Exit(code)
}

func TestSaveUser_Valid(t *testing.T) {
  p := NewPersist(&PersistProps{
    MongoSysDb: "test_boxtep_sys",
  })

  data := &User{
    Fullname: "King Long Too",
    Firstname: "Too",
    Username: "toolong",
    Password: "mypass",
    Email: "too@long.test",
  }
  SetRelaxed()
  user, err := p.SaveUser(context.TODO(), data)
  SetStrict()

  if err != nil {
    t.Fatalf("Should have saved user %#v", err)
  }

  data.Id = user.Id

  dataMap, _ := mongo.ToMap(data)
  userMap, _ := mongo.ToMap(user)

  mongo.AssertSubset(t, dataMap, userMap, "Should have been equal")
}

func TestSaveUser_RequireUsername(t *testing.T) {
  p := NewPersist(&PersistProps{
    MongoSysDb: "test_boxtep_sys",
  })

  data := &User{
    Fullname: "King Long Too",
    Firstname: "Too",
    Password: "mypass",
    Email: "too@long.test",
  }

  _, err := p.SaveUser(context.TODO(), data)

  if err == nil || err.Error() != `properties/Username: value "" too short for "minLength" argument 3` {
    t.Fatal("Should require username")
  }
}

func TestSaveUser_RequirePassword(t *testing.T) {
  p := NewPersist(&PersistProps{
    MongoSysDb: "test_boxtep_sys",
  })

  data := &User{
    Fullname: "King Long Too",
    Firstname: "Too",
    Username: "toolong",
    Email: "too@long.test",
  }

  _, err := p.SaveUser(context.TODO(), data)

  if err == nil || err.Error() != `properties/Password: value "" too short for "minLength" argument 1` {
    t.Fatal("Should require password")
  }
}

func TestSaveUser_UsesSysDb(t *testing.T) {
  db := "test_boxtep_sys"
  p := NewPersist(&PersistProps{
    MongoSysDb: db,
  })

  data := &User{
    Fullname: "King Long Too",
    Firstname: "Too",
    Username: "toolong",
    Password: "1",
    Email: "too@long.test",
  }

  _, err := p.SaveUser(context.TODO(), data)

  if err != nil {
    t.Fatalf("Should have saved user %#v", err)
  }

  // check if user is saved
  client := mongo.GetMongoClient()

  collection := client.GetCollection(db, USER_COLLECTION)

  var found bson.M
  filter := bson.M{"username": "toolong"}
  if err := collection.FindOne(context.Background(), filter).Decode(&found); err != nil {
    t.Fatalf("Should have found user %#v", err)
  }

  delete(found, "_id")
  delete(found, "password")
  delete(found, "createdAt")
	delete(found, "updatedAt")

  d, _ := mongo.ToMap(data)

  mongo.AssertSubset(t, found, d, "Should have been equal")
}

func TestAuthUser_UsingUsername(t *testing.T) {
  db := "test_boxtep_sys"
  p := NewPersist(&PersistProps{
    MongoSysDb: db,
  })

  data := &User{
    Fullname: "King Long Too",
    Firstname: "Too",
    Username: "toolong_test1",
    Password: "1",
    Email: "too@long.test",
  }

  _, err := p.SaveUser(context.TODO(), data)

  if err != nil {
    t.Fatalf("Should have saved user %#v", err)
  }

  user, err := p.CheckUser(context.TODO(), "toolong_test1", "1")
  if err != nil {
    t.Fatalf("Should have checked %#v", err)
  }

  d, _ := mongo.ToMap(data)
  u, _ := mongo.ToMap(user)

  delete(u, "_id")

  mongo.AssertSubset(t, u, d, "Wrong user")
}

func TestInc_First(t *testing.T) {
  SetRelaxed()
  defer SetStrict()

  db := "test_boxtep_sys"
  p := NewPersist(&PersistProps{
    MongoSysDb: db,
  })

  ctx := context.Background()

  v, err := p.Inc(ctx, "mycntr", 4)
  if err != nil {
    t.Fatalf("Should have increased: %v", err)
  }

  if v != 4 {
    t.Fatalf("Wrong number %d", v)
  }
}

func TestInc(t *testing.T) {
  SetRelaxed()
  defer SetStrict()

  db := "test_boxtep_sys"
  p := NewPersist(&PersistProps{
    MongoSysDb: db,
  })

  ctx := context.Background()
  p.Inc(ctx, "c1", 4)
  v, err := p.Inc(ctx, "c1", 1)
  if err != nil {
    t.Fatalf("Should have increased: %v", err)
  }

  if v != 5 {
    t.Fatalf("Wrong number %d", v)
  }
}

func TestBump(t *testing.T) {
  SetRelaxed()
  defer SetStrict()

  db := "test_boxtep_sys"
  p := NewPersist(&PersistProps{ MongoSysDb: db })

  ctx := context.Background()
  p.Bump(ctx, "c2")
  v := p.Bump(ctx, "c2")
  
  if v != 2 {
    t.Fatalf("Wrong number %d", v)
  }
}