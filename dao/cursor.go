package dao

import (
	"context"
	"encoding/json"
	"time"

	"cloud.google.com/go/datastore"
)

const (
	KindCursor = "cursor"
)

type Cursor struct {
	Cursors []byte    `datastore:",noindex"`
	Date    time.Time `datastore:",noindex"`

	K *datastore.Key `datastore:"__key__"`
}

type CursorDetails map[string]string

func CreateCursor(ctx context.Context, details CursorDetails) (string, error) {

	bb, err := json.Marshal(details)
	if err != nil {
		return "", err
	}

	token := newToken("c_", 10)

	c := &Cursor{
		Cursors: bb,
		Date:    time.Now(),
	}

	key := datastore.NameKey(KindCursor, token, nil)
	_, err = ds.Put(ctx, key, c)
	if err != nil {
		return "", err
	}

	return token, err
}

func ReadCursor(ctx context.Context, cursor string) (CursorDetails, error) {
	if cursor == "" {
		return map[string]string{}, nil
	}

	key := datastore.NameKey(KindCursor, cursor, nil)
	l := &Cursor{}
	err := ds.Get(ctx, key, l)
	if err != nil {
		return nil, err
	}
	c := map[string]string{}
	err = json.Unmarshal(l.Cursors, &c)

	return c, err
}
