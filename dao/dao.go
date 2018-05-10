package dao

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
)

var (
	ds *datastore.Client

	ErrNotFound = datastore.ErrNoSuchEntity
)

func Init() error {
	ctx, cf := context.WithTimeout(context.Background(), time.Second*10)
	defer cf()
	var err error
	ds, err = datastore.NewClient(ctx, "my-project")

	return err
}
