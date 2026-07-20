package structful

type IAdaptor interface {

	Save(data map[string]any) error

	GetWithFilter(filter map[string]any) ([]map[string]any, error)

	// GetSysDb returns the name of the database to use.
	GetDb() string

	// GetName the collection to store structful data.
	GetName() string

	CheckHash(string) bool

	List() ([]map[string]any, error)
}
