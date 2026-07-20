package structful

import (
	"log"

	"github.com/ianlancetaylor/jsonschema"
	"github.com/ianlancetaylor/jsonschema/draft7"
)

var validator *jsonschema.Schema

func init() {
	builder := draft7.NewBuilder()

	builder.AddType("object")
	
	builder.AddProperties(map[string]*jsonschema.Schema{
		"_group": draft7.NewSubBuilder().
			AddType("string").
			Build(), 
		"_name": draft7.NewSubBuilder().
			AddType("string").
			Build(),
		"_version": draft7.NewSubBuilder().
			AddType("string").
			Build(),
	})

	builder.AddRequired([]string{"_group", "_name", "_version"})

	builder.AddAdditionalProperties(draft7.NewSubBuilder().BoolSchema(true))

	validator = builder.Build()
}

var structful *Structful

type IStructful interface {
	Within(version string) *Structful

	Save(data map[string]any) error

	GetByGroup(group string) []map[string]any

	FilterByGroup(group string, filter map[string]any) []map[string]any

	GetOne(group, name string) map[string]any

	CheckHash(string) bool

	List() ([]map[string]any, error)
}

type Structful struct {
	IStructful

	SysDb string

	Name string

	Version string

	Adaptor IAdaptor
}

func CreateStructful(props *StructfulProps) *Structful {
	if props.Adaptor == AdaptorType_Mongo {
		s := &Structful{
			Adaptor: CreateMongoAdaptor(props.SysDb, props.Name),
			Version: props.Version,
			Name: props.Name,
			SysDb: props.SysDb,
		}

		return s
	} else {
		log.Fatal("unsuported adaptor")

		return nil
	}
}

type AdaptorType int

const (
	AdaptorType_Mongo = iota
)

func(s *Structful) Save(data map[string]any) error {
	if err := validator.Validate(data); err != nil {
		return err
	}

	return s.Adaptor.Save(data)
}

func (s *Structful) FilterByGroup(group string, filter map[string]any) ([]map[string]any, error) {
	filter["_group"] = group

	return s.Adaptor.GetWithFilter(filter)
}

func (s *Structful) GetByGroup(group string) ([]map[string]any, error) {
	return s.FilterByGroup(group, map[string]any{})
}

func (s *Structful) GetOne(group, name string) (map[string]any, error) {
	filter := map[string]any{ "_name": name }
	
	results, err := s.FilterByGroup(group, filter)
	if err != nil {
		return nil, err
	}

	if len(results) > 0 {
		return results[0], nil
	} else {
		return nil, nil
	}
}

func (s *Structful) CheckHash(hash string) bool {
	return s.Adaptor.CheckHash(hash)
}

func (s *Structful) List() ([]map[string]any, error) {
	return s.Adaptor.List()
}

func Current() *Structful {
	return structful
}

type StructfulProps struct {
	Adaptor AdaptorType
	SysDb string
	Name string
	Version string
}

func GetStructful(props *StructfulProps) *Structful {
	if props.Adaptor == AdaptorType_Mongo {
		return &Structful{
			Adaptor: CreateMongoAdaptor(props.SysDb, props.Name),
			SysDb: props.SysDb,
			Version: props.Version,
			Name: props.Name,
		}
	} else {
		return nil
	}
}

type InitProps struct {
	MongoSysDb string // default: db_sys
	Version string // default: 0
	MongoStructfulCollection string // default: _structful
}

func Init(props *InitProps) {
	sysDb	:= props.MongoSysDb
	version := props.Version
	name := props.MongoStructfulCollection

	structful = GetStructful(&StructfulProps{
		Adaptor: AdaptorType_Mongo,
		SysDb: sysDb,
		Version: version,
		Name: name,
	})
}
