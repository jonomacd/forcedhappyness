package dao

import (
	"context"

	"cloud.google.com/go/datastore"
	"golang.org/x/crypto/acme/autocert"
)

type autocertCache struct {
	ds     *datastore.Client
	kind   string
	prefix string
}

type cert struct {
	Cert []byte `datastore:",noindex"`
}

func NewCertCache(prefix string) autocert.Cache {
	return autocertCache{
		ds:     ds,
		kind:   "KindCertCache",
		prefix: prefix,
	}
}

func (ac autocertCache) Get(ctx context.Context, name string) ([]byte, error) {
	key := datastore.NameKey(ac.kind, ac.prefix+name, nil)
	c := &cert{}
	err := ds.Get(ctx, key, c)
	if err != nil {
		return nil, err
	}
	return c.Cert, nil
}

func (ac autocertCache) Put(ctx context.Context, name string, data []byte) error {
	c := &cert{
		Cert: data,
	}

	key := datastore.NameKey(ac.kind, ac.prefix+name, nil)
	_, err := ds.Put(ctx, key, c)
	return err
}

func (ac autocertCache) Delete(ctx context.Context, name string) error {
	key := datastore.NameKey(ac.kind, ac.prefix+name, nil)
	return ds.Delete(ctx, key)
}
