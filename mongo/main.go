// Package mongo provides a simple interface to the database.
package mongo

import (
	"context"
	"log"
	"slices"
	"strings"
	"time"
	"fmt"
	"encoding/json"
	"sync"

	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gsuntres/pkg/commons"
	"github.com/gsuntres/pkg/sys"
	"github.com/gsuntres/pkg/events"
)

var mux sync.RWMutex

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// WithiSessionFunc will run operations on the database within the same session.
// If this functions returns an error, the system will rollback the transaction.
// Sessions require a resplica set.
type WithinSessionFunc func(context.Context, *mongo.Session) (any, error)

type IMongoClient interface {
	// WithinSession takes a context, the database and the collection and excecutes the callback function provided.
	WithinSession(context.Context, WithinSessionFunc) (any, error)

	// InsertOne will insert data to the specified namespace.
	InsertOne(context.Context, string, string, bson.M) (bson.M, error)

	// Find
	Find(context.Context, string, string, bson.M) ([]bson.M, error)

	// GetCollection returns the requested collection within an account. It will create it if it doesn't exist.
	GetCollection(database, name string) *mongo.Collection

	SetRelaxed()

	SetStrict()
}

type Timeseries struct {
	TimeField     string `bson:"timeField"`
	MetaField     string `bson:"metaField"`
	Granularity   string `bson:"granularity"`
	BucketMaxSpan int64  `bson:"bucketMaxSpan"`
}

type Discriminator struct {
	Field         string `bson:"field"`
	CtxField      string `bson:"ctxField"`
	Collection    string `bson:"collection"`
	Required      bool   `bson:"required,omitempty"`
}

type CollectionDefinition struct {
	Name              string 		   `bson:"_name"`
	Singular          string 		   `bson:"singular"`
	Plural            string 		   `bson:"plural"`
	IdPrefix          string 		   `bson:"idPrefix"`
	IndexSpecs    	  []map[string]any `bson:"indexSpecs"`
	Schema            map[string]any   `bson:"schema"`
	DisableValidation bool 			   `bson:"disableValidation"`
	Views 	          map[string]any   `bson:"views"`
	Ttl               int64 		   `bson:"ttl"`
	Timeseries        *Timeseries      `bson:"timeseries,omitempty"`
	Discriminator     *Discriminator   `bson:"discriminator"`
}

// MongoClient
type MongoClient struct {
	// Client the actual connected instance of mongo client.
	Client *mongo.Client
	// Debug set to true for dislaying info level logs.
	Debug bool
	// DebugQuery set to true to log all queries done.
	DebugQuery bool
	// Limit the default limit to use in queries.
	Limit int64
	// DBPrefix optinal string literal to distinquise
	DBPrefix string
	// Registry holds critical information about collection's structure, like schema and indexes
	Registry map[string]*CollectionDefinition
	// Alias collection names that map to a registered collection
	Aliases map[string]string
	// Relaxed if set to true will not enforce schema
	Relaxed bool
	// Context fields to serialize when needed
	ContextFields []string
	// WithAudit enable auditing functionality.
	WithAudit bool
	// OnAudit callback called when audit is enabled.
	OnAudit *OnAudit
	// EventBus
	EventBus *events.TypedEventBus
	// IgnoreAudit a list of entities to ignore
	IgnoreAudit []string

	flashDb string
}

func (c *MongoClient) HasFlash() bool {
	mux.Lock()
	defer mux.Unlock()

	return c.flashDb != ""
}

func (c *MongoClient) Flash(db string) {
	mux.Lock()
	defer mux.Unlock()
	c.flashDb = db
}

func (c *MongoClient) GetFlash() string {
	mux.Lock()
	defer mux.Unlock()
	defer (func() { c.flashDb = "" })()

	return c.flashDb
}

// AddAlias will add the alias only if it doesn't exist.
func (c *MongoClient) AddAlias(alias, name string) {
	if c.Aliases == nil {
		c.Aliases = make(map[string]string)
	}
	_, ok := c.Aliases[alias]
	if !ok {
		log.Printf("Registering alias %s to %s", alias, name)
		c.Aliases[alias] = name
	}
}

func (c *MongoClient) SetRelaxed() {
	c.Relaxed = true
}

func (c *MongoClient) SetStrict() {
	c.Relaxed = false
}

func (c *MongoClient) GetIdPrefix(name string) string {
	def, ok := c.Registry[name]
	if ok {
		return def.IdPrefix
	}

	return ""
}

func (c *MongoClient) AddDefinition(data map[string]any) {
	if valid := commons.Validate(ADD_DEFINITION_SCHEMA, data); valid != nil {
		log.Printf("failed to register data: %v", valid)

		return
	}

	log.Printf("Registering %s", data["_name"])

	b, err := bson.Marshal(data)
	if err != nil {
		log.Printf("failed to marshal: %v", err)

		return
	}

	var cd CollectionDefinition
	bson.Unmarshal(b, &cd)

	if len(cd.Name) > 0 {
		c.Registry[cd.Name] = &cd
	}

	if cd.Views != nil {
		for k, _ := range cd.Views {
			c.AddAlias(k, cd.Name)
		}
	}
}

func (c *MongoClient) DropDatabase_DANGER(database string) bool {
	log.Printf("DANGER attempt to drop database: %s", database)

	db := c.Client.Database(database)

	if err := db.Drop(context.Background()); err != nil {
		log.Printf("Failed to drop database: %v", err)

		return false
	}

	log.Printf("Database %s deleted", database)

	return true
}

func (c *MongoClient) GetCollection(database, name string) *mongo.Collection {
	if c.Debug {
		log.Printf("Using collection: %s.%s", database, name)
	}

	db := c.Client.Database(database)

	// To make sure we create the right collection when dealing with aliases.
	actualName := name

	return db.Collection(actualName)
}

func (c *MongoClient) GetCollectionWithCheck(database, name string) *mongo.Collection {
	if c.Debug {
		log.Printf("Using collection: %s.%s", database, name)
	}

	db := c.Client.Database(database)

	// To make sure we create the right collection when dealing with aliases.
	actualName := name

	// 1. List existing collections
	names, err := db.ListCollectionNames(context.TODO(), bson.D{})

	if err != nil {
		log.Printf("Failed to list collections: %#v", err.Error())

		return nil
	}

	// 2. If collection exist return it, otherwise create it and then return it
	if slices.Contains(names, actualName) {
		return db.Collection(actualName)
	} else {
		opts := options.CreateCollection()

		// maybe get from schema
		cdef, ok := c.Registry[actualName]
		if ok {
			actualName = cdef.Name

			log.Printf("Schema found for %s; will use it", actualName)

			ApplyTimeSeries(cdef, opts)
			ApplySchema(cdef, opts)
		} else {
			log.Printf("No schema for %s", actualName)

			if c.Relaxed == false {
				return nil	
			}
		}

		if err := db.CreateCollection(context.TODO(), actualName, opts); err != nil {
			log.Printf("Failed to create collection: %#v", err)

			return nil
		}

		collection := db.Collection(actualName)
		c.CreateIndexes(collection, cdef)

		if c.CreateViews(db, cdef) {
			collection = db.Collection(name)
		}

		return collection
	}
}

func ApplyTimeSeries(cdef *CollectionDefinition, opts *options.CreateCollectionOptionsBuilder) {
	if cdef.Timeseries != nil {
		tsOpts := options.TimeSeries().
			SetTimeField(cdef.Timeseries.TimeField).
			SetMetaField(cdef.Timeseries.MetaField).
			SetGranularity(cdef.Timeseries.Granularity)

		if cdef.Timeseries.BucketMaxSpan > 0 {
			tsOpts.SetBucketMaxSpan(time.Duration(cdef.Timeseries.BucketMaxSpan) * time.Second)
		}

		opts.SetTimeSeriesOptions(tsOpts)

		opts.SetExpireAfterSeconds(cdef.Ttl)
	} else {
		log.Printf("No timeseries")
	}
}

func ApplySchema(cdef *CollectionDefinition, opts *options.CreateCollectionOptionsBuilder) {
	// Add schema validation
	if cdef.Schema != nil && cdef.Timeseries == nil && cdef.DisableValidation == false {
		schemaBson, err := bson.Marshal(cdef.Schema)
		if err != nil {
			log.Printf("failed to parse schema: %v", err)
		} else {
			var validatorSchema bson.M
			bson.Unmarshal(schemaBson, &validatorSchema)
			validator := bson.M{
				"$jsonSchema": validatorSchema,
			}

			opts.SetValidator(validator)
		}
	} else {
		log.Printf("Validation disabled for %s", cdef.Name)
	}
}

var client *MongoClient = &MongoClient{
	Limit: 10,
	Registry: make(map[string]*CollectionDefinition, 0),
	Relaxed: false,
	IgnoreAudit: []string{"event"},
}

func GetMongoClient() *MongoClient {
	return client
}

const validSchema_StartProps = `
{
	"type": "object",
	"properties": {
		"MongoUri": {
			"type": "string"
		},
		"MongoUser": {
			"type": "string"
		},
		"MongoPass": {
			"type": "string"
		}
	},
	"required": ["MongoUri"]
}
`

type MongoStartProps struct {
	MongoUri string
	MongoUser string
	MongoPass string
	MongoDebugQuery bool
	MongoDBPrefix string
	ContextFields []string
	WithAudit bool
	NoAutoCleanup bool
	OnAudit *OnAudit
	AllowTruncatingDoubles bool
}

func Start(props *MongoStartProps) error {
	if err := commons.Validate(validSchema_StartProps, props); err != nil {
		return err
	}

	uri := props.MongoUri
	user := props.MongoUser
	pass := props.MongoPass

	// Create a new client and connect to the server
	var err error

	cOptions := options.Client().
		ApplyURI(uri).
		SetAuth(options.Credential{
			Username: user,
			Password: pass,
		}).
		SetConnectTimeout(5 * time.Second).
		SetServerAPIOptions(&options.ServerAPIOptions{
			ServerAPIVersion: options.ServerAPIVersion1,
		}).
		SetBSONOptions(&options.BSONOptions{
			DefaultDocumentM:  true,
			UseJSONStructTags: false,
			NilSliceAsEmpty:   true,
        	NilMapAsEmpty:     true,
        	AllowTruncatingDoubles: props.AllowTruncatingDoubles,
		}).
		SetRegistry(GetCustomRegistry())

	colors := []string{
		commons.EscapeGreen,
		commons.EscapeRed,
		commons.EscapeYellow,
		commons.EscapeBlue,
		commons.EscapeMagenta,
		commons.EscapeCyan,
	}

	client.DebugQuery = props.MongoDebugQuery
	if client.DebugQuery {
		// Debug queries
		monitor := &event.CommandMonitor{
			Started: func(_ context.Context, e *event.CommandStartedEvent) {
				ecode := colors[e.RequestID % 6]
				log.Printf("%s%d@Start%s %s#%s %s", ecode, e.RequestID, commons.EscapeReset, e.DatabaseName, e.CommandName, e.Command)
			},
			Succeeded: func(_ context.Context, e *event.CommandSucceededEvent) {
				ecode := colors[e.RequestID % 6]
				log.Printf("%s%d@OK%s in %s", ecode, e.RequestID, commons.EscapeReset, e.Reply)
			},
			Failed: func(_ context.Context, e *event.CommandFailedEvent) {
				ecode := colors[e.RequestID % 6]
				log.Printf("%s%d@Fail%s in %s", ecode, e.RequestID, commons.EscapeReset, e.Failure)
			},
		}

		cOptions = cOptions.SetMonitor(monitor)
	}

	// set DBPrefix
	client.DBPrefix = props.MongoDBPrefix

	client.ContextFields = props.ContextFields

	client.WithAudit = props.WithAudit

	client.EventBus = events.NewTypedEventBus()

	if client.WithAudit {
		var onAudit OnAudit = func(audit *AuditResult) error {
			return events.Publish(client.EventBus, audit)
	 	}
	 	client.OnAudit = &onAudit
	}
	
	if err := cOptions.Validate(); err != nil {
		log.Fatalf("Failed to validate mongo options: %+v", err.Error())

		return err
	}

	client.Client, err = mongo.Connect(cOptions)

	if err != nil {
		log.Fatalf("Failed to connect: %+v", err.Error())

		return err
	}

	if !props.NoAutoCleanup {
		sys.OnExit(func () {
			if err := Stop(); err != nil {
				log.Fatalf("failed to disconnect: %v", err)
			}
		})	
	}

	// Register defaults.go
	var structfulDef bson.M
	json.Unmarshal(structfulJson, &structfulDef)

	client.AddDefinition(structfulDef)

	return nil
}

func (c *MongoClient) Subscribe(h events.Handler[AuditResult]) (cancel func()) {
	return events.Subscribe(c.EventBus, h)
}

func GetClient() *mongo.Client {
	return client.Client
}

func (c *MongoClient) GetName(account string) string {
	return fmt.Sprintf("%s%s", c.DBPrefix, account)
}

func (c *MongoClient) ExtractAccount(name string) string {
	v, ok := strings.CutPrefix(name, c.DBPrefix)

	if !ok {
		return ""
	}

	return v
}

func Stop() error {
	if err := client.Client.Disconnect(context.TODO()); err != nil {
		return err
	}

	log.Printf("Successfully stopped mongo.")

	return nil
}

func MapToBsonD(m map[string]any) (bson.D, error) {
    data, err := bson.Marshal(m)
    if err != nil {
    	return nil, err
    }

    var d bson.D

	err = bson.Unmarshal(data, &d)
    if err != nil {
    	return nil, err
    }

    return d, nil
}

const ADD_DEFINITION_SCHEMA = `
{
	"type": "object",
	"properties": {
		"_name": {
			"type": "string"
		},
		"idPrefix": {
			"type": "string"
		},
		"indexSpecs": {
			"type": "array"
		},
		"singular": {
			"type": "string"
		},
		"system": {
			"type": "boolean"
		},
		"plural": {
			"type": "string"
		},
		"schema": {
			"type": "object"
		},
		"views": {
			"type": "object"
		},
		"ttl": {
			"type": "integer",
  			"format": "int64"
		},
		"timeseries": {
			"type": "object",
			"properties": {
				"timeField": {
					"type": "string"
				},
				"metaField": {
					"type": "string"
				},
				"granularity": {
					"enum": [ "seconds", "minutes", "hours"]
				},
				"bucketMaxSpan": {
					"type": "integer",
  					"format": "int64"
				}
			},
			"required": ["timeField", "metaField", "granularity"]
		},
		"discriminator": {
			"type": "object",
			"properties": {
				"field": {
					"type": "string"
				},
				"ctxField": {
					"type": "string"
				},
				"collection": {
	        		"type": "string"
	        	},
	        	"required": {
	        		"type": "boolean"
	        	}
			},
			"required": ["field", "ctxField", "collection"]
        }
	},
	"required": ["_name", "singular", "plural"],
	"additionalProperties": true
}
`