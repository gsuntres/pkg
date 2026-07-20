package options

type DeleteAudit struct {
	Id string
	Before any
	Context any
}

type OnDeleteAudit func(audit *DeleteAudit)

type DeleteOneOptions struct {
	OnDeleteAudit *OnDeleteAudit
}

type DeleteOneOptionsBuilder struct {
	Opts []func(*DeleteOneOptions) error
}

func DeleteOne() *DeleteOneOptionsBuilder{
	return &DeleteOneOptionsBuilder{}
}

func (dao *DeleteOneOptionsBuilder) List() []func(*DeleteOneOptions) error {
	return dao.Opts
}

func (dao *DeleteOneOptionsBuilder) WithOnDeleteAudit(oa *OnDeleteAudit) *DeleteOneOptionsBuilder {
	dao.Opts = append(dao.Opts, func(opts *DeleteOneOptions) error {
		opts.OnDeleteAudit = oa

		return nil
	})

	return dao
}