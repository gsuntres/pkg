package structful

import (
	"log"
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
	mongov2 "go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/gsuntres/pkg/mongo"
)

type MongoAdaptor struct {
	IAdaptor

	sysDb string

	name string
}

func CreateMongoAdaptor(sysDb string, name string) *MongoAdaptor {
	return &MongoAdaptor{
		sysDb: sysDb,
		name: name,
	}
}

func (a *MongoAdaptor) Save(data map[string]any) error {
	client := mongo.GetMongoClient()

	collection := client.GetCollection(a.GetDb(), a.GetName())

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		log.Fatalf("failed to insert %+v", err)
	}

	log.Printf("Inserted %+v", res)

	return nil
}

func (a *MongoAdaptor) CheckHash(hash string) bool {
	client := mongo.GetMongoClient()

	collection := client.GetCollection(a.GetDb(), a.GetName())

	opts := options.Find()

	filter := map[string]any{ "_hash": hash }
	cursor, err := collection.Find(context.Background(), filter, opts)
	if err != nil {
		log.Printf("ERROR %v", err)

		return true
	}

	return cursor.Next(context.Background())
}

func (a *MongoAdaptor) GetWithFilter(filter map[string]any) ([]map[string]any, error) {
	client := mongo.GetMongoClient()

	collection := client.GetCollection(a.GetDb(), a.GetName())

	opts := options.Find()

	var err error
	var cursor *mongov2.Cursor
	cursor, err = collection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}

	results := make([]map[string]any, 0)
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (a *MongoAdaptor) List() ([]map[string]any, error) {
	client := mongo.GetMongoClient()

	collection := client.GetCollection(a.GetDb(), a.GetName())

	opts := options.Find()

	var err error
	var cursor *mongov2.Cursor
	cursor, err = collection.Find(context.Background(), nil, opts)
	if err != nil {
		return nil, err
	}

	results := make([]map[string]any, 0)
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (a *MongoAdaptor) GetDb() string {
	return a.sysDb
}

func (a *MongoAdaptor) GetName() string {
	return a.name
}
