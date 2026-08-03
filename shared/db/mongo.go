package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewMongo(isDev bool, uri, db string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Client().ApplyURI(uri).
		SetRetryWrites(true).
		SetRetryReads(true).
		SetConnectTimeout(10 * time.Second)
	if isDev {
		opts.SetMaxPoolSize(4)
	} else {
		opts.
			SetMaxPoolSize(50).
			SetMinPoolSize(10)
	}
	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}
