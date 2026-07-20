package options

type FindOneOptions struct {
	Projection any
}

type FindOneOptionsFunc func(*FindOneOptions) error

type FindOneOptionsBuilder struct {
	Opts []FindOneOptionsFunc
}

func FindOne() *FindOneOptionsBuilder {
	return &FindOneOptionsBuilder{
    Opts: make([]FindOneOptionsFunc, 0),
  }
}

func (builder *FindOneOptionsBuilder) SetProjection(projection any) *FindOneOptionsBuilder {
  builder.Opts = append(builder.Opts, func(options *FindOneOptions) error {
    options.Projection = projection
    
    return nil
  })

  return builder
}