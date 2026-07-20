package mongo

import (
	"log"
	"fmt"
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gsuntres/pkg/commons"
)

func (c *MongoClient) DiscriminatorCheckAndApplyToData(ctx context.Context, name string, data map[string]any) error {
	cdef, ok := c.Registry[name]
	if ok && cdef.Discriminator != nil {
		if data == nil {
			data = map[string]any{}
		}

		log.Printf("Discriminator found for %s; will use it", name)

		// get from context
		vAny := ctx.Value(cdef.Discriminator.CtxField)
		if vAny == nil {
			return fmt.Errorf("discriminator field required for %s", name)
		}

		// update payload
		v := vAny.(string)
		data[cdef.Discriminator.Field] = v
	}

	return nil
}

func (c *MongoClient) DiscriminatorOmitInData(name string, data bson.M) error {
	cdef, ok := c.Registry[name]
	if ok && cdef.Discriminator != nil {
		if data == nil {
			data = map[string]any{}
		}

		log.Printf("Making sure discriminator is not in data for %s", name)

		_, ok := data[cdef.Discriminator.Field]
		if ok {
			delete(data, cdef.Discriminator.Field)
		}
	}

	return nil
}

type DiscriminatorResult struct {
	WithDiscriminator   bool
	DiscriminatorField string
}

func (c *MongoClient) DiscriminatorCheckAndApplyToFilter(ctx context.Context, name string, filter bson.M) (*DiscriminatorResult ,error) {
	result := &DiscriminatorResult{}

	cdef, ok := c.Registry[name]
	if ok && cdef.Discriminator != nil {
		if filter == nil {
			filter = bson.M{}
		}

		log.Printf("Discriminator found for %s; will use it", name)

		// get from context
		vAny := ctx.Value(cdef.Discriminator.CtxField)
		if vAny == nil {
			return nil, fmt.Errorf("discriminator field required for %s", name)
		}

		// update payload
		v := vAny.(string)
		if commons.StringIsBlank(v) {
			return nil, fmt.Errorf("discriminator field required for %s", name)
		}

		filter[cdef.Discriminator.Field] = bson.M{"eq": v}

		result.DiscriminatorField = cdef.Discriminator.Field
		result.WithDiscriminator = true
	}

	return result, nil
}

