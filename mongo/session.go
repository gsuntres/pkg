package mongo

import (
	"fmt"
	"errors"
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)


func (c *MongoClient) WithinSession(ctx context.Context, cb WithinSessionFunc) (any, error) {
	
	// 1. Start a new session
	sess, err := c.Client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}

	defer sess.EndSession(ctx)

	// 2. Create a new context
	ctxNew := mongo.NewSessionContext(ctx, sess)

	// 3. Start transaction
	if err = sess.StartTransaction(); err != nil {
		return nil, fmt.Errorf("start transaction: %w", err)
	}

	res, err := cb(ctxNew, sess)
	if err != nil {
		errWrap := fmt.Errorf("transaction callback failed: %w", err)
		if errAbort := sess.AbortTransaction(ctxNew); errAbort != nil {
			return nil, errors.Join(
				errWrap,
        fmt.Errorf("failed to abort transaction: %w", errAbort),
      )
		}

		return nil, errWrap
	}

	if err := sess.CommitTransaction(ctxNew); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return res, nil
}