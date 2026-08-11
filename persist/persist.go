// Package persist offers a convinient way to interact with the database. It takes into consideration multitenancy.
package persist

import (
	"log"
	"context"
	"fmt"
	"errors"
	"reflect"
	
	v2mongo "go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/go-viper/mapstructure/v2"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gsuntres/pkg/commons"
	"github.com/gsuntres/pkg/mongo"
	"github.com/gsuntres/pkg/structful"
	"github.com/gsuntres/pkg/events"
)

type PersistProps struct {
	MongoSysDb string	
}

// Persist is the root object for accessing data.
type Persist struct {
	// IPersist

	SysDb string
}

var fields []reflect.StructField

var sysDb string

func NewPersist(props *PersistProps) *Persist {
	p := &Persist{}

	p.SysDb = props.MongoSysDb
	
	return p
}

var persistOriginal *Persist

var deferedFuncs map[string]any = make(map[string]any, 0)

var persist any

type InitReport struct {
	Fields []string
}

func (report *InitReport) AddField(field string) {
	report.Fields = append(report.Fields, field)
}

type InitProps struct {
	MongoSysDb string
}

// Init runs on boot, reads system_collections from structful and registers them. No changes after boot are going to be applied.
func Init(props *InitProps) {
	var persistProps PersistProps
	commons.StructToStruct(props, &persistProps)
	persist = NewPersist(&persistProps)

	if props.MongoSysDb != "" {
		sysDb = props.MongoSysDb	
	} else {
		sysDb = "sys_db"
	}
	
	persistOriginal = NewPersist(&persistProps)

	// Load structful
	s := structful.Current()

	filter := map[string]any{}
	scol, err := s.FilterByGroup("system_collections", filter)
	if err != nil {
		log.Fatalf("Failed to get system collections: %v", err)
	}

	// TODO(me): WIP
	// Build a new struct similar to Persist

	origType := reflect.TypeOf(persist)
	if origType.Kind() == reflect.Ptr {
		origType = origType.Elem()
	}

	// 1. Copy all existing fields
	for i := 0; i < origType.NumField(); i++ {
		// log.Printf("Add existing field: %v", origType.Field(i))
		fields = append(fields, origType.Field(i))
	}

	client := mongo.GetMongoClient()

	report := &InitReport{}

	// Take system_collection from structful and build the calls.
	for _, col := range scol {
		BuildInsertOne(col, report)
		BuildFindOne(col, report)
		BuildFind(col, report)
		BuildFindOffset(col, report)
		BuildGetOne(col, report)
		BuildDeleteOne(col, report)
		BuildReplaceOne(col, report)
		BuildUpdateSet(col, report)
		BuildUpdateAddSet(col, report)
		BuildCount(col, report)

		client.AddDefinition(col)
	}

	BuildFindGeneric()

	log.Printf("Fields Registered: %v", report.Fields)

	// 3. Create the new struct type
	newStructType := reflect.StructOf(fields)

	// 4. Create a new instance
	newStruct := reflect.New(newStructType).Elem()

	for _, field := range fields {
		// log.Printf("FIELD %s", field.Name)
		if actualFunc, ok := deferedFuncs[field.Name]; ok {
			f :=	newStruct.FieldByName(field.Name)
			f.Set(reflect.ValueOf(actualFunc))
		}
	}

	persist = newStruct.Addr().Interface()
}

func GetCurrent() any {
	return persist
}

func Orig() *Persist {
	return persistOriginal
}

func (p *Persist) Subscribe(h events.Handler[mongo.AuditResult]) (cancel func()) {
	client := mongo.GetMongoClient()
	return client.Subscribe(h)
}

func (p *Persist) DisableAudit() {
	client := mongo.GetMongoClient()
	client.WithAudit = false
}

func (p *Persist) EnableAudit() {
	client := mongo.GetMongoClient()
	client.WithAudit = true
}

func (p *Persist) BootAccount(account string) error {
	// Load structful
	s := structful.Current()

	filter := map[string]any{}
	scol, err := s.FilterByGroup("system_collections", filter)
	if err != nil {
		log.Fatalf("Failed to get system collections: %v", err)
	}

	client := mongo.GetMongoClient()

	log.Printf("Booting account %s", account)
	for _, n := range scol {
		name := fmt.Sprintf("%s", n["_name"])
		
		log.Printf("boot collection %s", name)

		// 1. Check if system
		isSystem := false
		if v, ok := n["system"]; ok {
			isSystem = v.(bool)
		}

		var db string
		if isSystem {
			db = sysDb
		} else {
			db = client.GetName(account)
		}

		client.GetCollectionWithCheck(db, name)
	}

	return nil
}

var ErrCall error = errors.New("failed to call")

type WithinSessionFunc func(context.Context, any) (any, error)

func WithinSession(ctx context.Context, cb WithinSessionFunc) (any, error) {
	client := mongo.GetMongoClient()

	return client.WithinSession(ctx, func(sessCtx context.Context, sess *v2mongo.Session) (any, error) {
		p := GetCurrent()
		return cb(sessCtx, p)
	})
}

func Call(p any, name string, args... any) ([]reflect.Value, error) {
	v := reflect.ValueOf(p)

	// 1. Make sure it's the right type
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// 2. Get field
	fn := v.FieldByName(name)

	// 3. Validate
	if !fn.IsValid() {
		log.Printf("Failed to find %s field", name)

		return nil, ErrCall
	}

	if fn.Kind() != reflect.Func {
		log.Printf("%s is not a function", name)

		return nil, ErrCall
	}

	if fn.IsNil() {
		log.Printf("%s is nil", name)

		return nil, ErrCall
	}

	// 4. Prepare arguments
	vArgs := make([]reflect.Value, len(args))

	for i, arg := range args {
  	vArgs[i] = reflect.ValueOf(arg)
  }

	// 5. Call and return results
	return fn.Call(vArgs), nil
}

const SAVE_USER_DATA = `
{
	"type": "object",
	"properties": {
		"Fullname": {
			"type": "string"
		},
		"Firstname": {
			"type": "string"
		},
		"Username": {
			"type": "string",
			"minLength": 3
		},
		"Password": {
			"type": "string",
			"minLength": 1
		},
		"Email": {
			"type": "string"
		}
	},
	"required": ["Username", "Password"]
}
`

// SaveUser will encrypt the password before saving the user.
func (p *Persist) SaveUser(ctx context.Context, data *User) (*User, error) {
	if valid := commons.Validate(SAVE_USER_DATA, data); valid != nil {
		return nil, errors.New(fmt.Sprintf("%s", valid.Error()))
	}

	data.EncryptPassword()

	client := mongo.GetMongoClient()

	dataNew, err := client.InsertOneFromStruct(ctx, p.SysDb, USER_COLLECTION, data)
	if err != nil {
		return nil, err
	}
	d, err := bson.Marshal(dataNew)
	if err != nil {
		return nil, err
	}

	// Unmarshal into struct
	var u User
	err = bson.Unmarshal(d, &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (p *Persist) UpdatePassword(ctx context.Context, username string, oldPass, newPass string) error {
	if commons.StringIsBlank(oldPass) || commons.StringIsBlank(newPass) {
		return errors.New("invalid password (blank)")
	}

	if commons.StringIsBlank(username) {
		return errors.New("invalid username")
	}

	user, err := p.CheckUser(ctx, username, oldPass)
	if err != nil {
		return err
	}

	user.Password = newPass
	user.EncryptPassword()

	client := mongo.GetMongoClient()
	
	_, err = client.UpdateSet(ctx, p.SysDb, USER_COLLECTION, bson.M{"username": username}, bson.M{"password": user.Password})
	if err != nil {
		return err
	}

	return nil
}

func (p *Persist) Bump(ctx context.Context, name string) int64 {
	v, err := p.Inc(ctx, name, int64(1))
	if err != nil {
		log.Fatalf("Failed to bump counter: %v", err)
	}

	return v
}

func (p *Persist) Inc(ctx context.Context, name string, v int64) (int64, error) {
	var zero int64

	mc := mongo.GetMongoClient()

	// figure out what database to use
	var db string
	var account string
	accountAny := ctx.Value("account")
	if accountAny != nil {
		account = accountAny.(string)
	}
	if account != "" {
		db = mc.GetName(account)
	} else {
		db = p.SysDb
	}

	filter := bson.M{
		"type": "counter",
		"name": name,
	}

	inc := bson.M{
		"counter": v,
	}

	o, err := mc.UpdateInc(ctx, db, "singleton", filter, inc)
	if err != nil {
		return zero, err
	}

	next, ok := commons.GetGeneric[int64](o, "counter")
	if !ok {
		return zero, errors.New("counter not found")
	}

	return next, nil
}

func (p *Persist) HasFlash() bool {
	client := mongo.GetMongoClient()

	return client.HasFlash()
}

func (p *Persist) Flash(db string) {
	client := mongo.GetMongoClient()

	client.Flash(db)
}

func (p *Persist) GetFlash() string {
	client := mongo.GetMongoClient()

	return client.GetFlash()
}

func (p *Persist) CheckUser(ctx context.Context, usernameOrEmail string, password string) (*User, error) {
	if len(usernameOrEmail) < 3 {
		return nil, errors.New("username or email too short")
	}

	client := mongo.GetMongoClient()

	filter := bson.M{
		"$or": bson.A{
			bson.M{"username": usernameOrEmail},
			bson.M{"email": usernameOrEmail},
		},
	}

	found, err := client.FindOne(ctx, p.SysDb, USER_COLLECTION, filter)
	if err != nil {
		return nil, err
	}

	var user User
	if err := mongo.ToStruct(found, &user); err != nil {
		return nil, err
	}

	if !user.CheckPassword(password) {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

const SAVE_ACCOUNT_DATA = `
{
	"type": "object",
	"properties": {
		"Code": {
			"type": "string",
			"minLength": 6
		},
		"Owner": {
			"type": "string",
			"minLength": 6
		}
	},
	"required": ["Code", "Owner"]
}
`

func (p *Persist) SaveAccount(ctx context.Context, data *Account) (*Account, error) {
	if valid := commons.Validate(SAVE_ACCOUNT_DATA, data); valid != nil {
		return nil, errors.New(fmt.Sprintf("%s", valid.Error()))
	}

	client := mongo.GetMongoClient()

	dataNew, err := client.InsertOneFromStruct(ctx, p.SysDb, ACCOUNT_COLLECTION, data)
	if err != nil {
		return nil, err
	}

	var o Account
	mapstructure.Decode(dataNew, &o)

	return &o, nil
}

func (p *Persist) GetAccountByCode(ctx context.Context, code string) (*Account, error) {
	client := mongo.GetMongoClient()

	filter := bson.M{"code": code}
	found, err := client.FindOne(ctx, p.SysDb, ACCOUNT_COLLECTION, filter)
	if err != nil {
		return nil, err
	}

	var acc Account
	if err := mongo.ToStruct(found, &acc); err != nil {
		return nil, err
	}

	return &acc, nil
}

func SetRelaxed() {
	mongo.GetMongoClient().SetRelaxed()
}

func SetStrict() {
	mongo.GetMongoClient().SetStrict()
}