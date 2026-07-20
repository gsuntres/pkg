package options

type Lister[T any] interface {
	List() []func(*T) error
}