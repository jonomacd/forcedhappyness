package dao

import (
	"context"

	"cloud.google.com/go/datastore"
	"github.com/jonomacd/forcedhappyness/site/domain"
)

const (
	KindPost = "post"
)

type Post struct {
	domain.Post `datastore:",flatten"`

	K *datastore.Key `datastore:"__key__"`
}

func CreatePost(ctx context.Context, p Post) error {
	if p.ID == "" {
		p.ID = newToken("p_", 15)
	}
	key := datastore.NameKey(KindPost, p.ID, nil)
	_, err := ds.Put(ctx, key, &p)
	return err
}

func ReadPostByID(ctx context.Context, id string) (Post, error) {
	key := datastore.NameKey(KindPost, id, nil)
	p := Post{}
	err := ds.Get(ctx, key, &p)
	return p, err
}

func ReadPostsByTime(ctx context.Context, offset, limit int) ([]Post, error) {
	q := datastore.NewQuery(KindPost).Order("-Date").Offset(offset).Limit(limit)

	posts := []Post{}
	_, err := ds.GetAll(ctx, q, &posts)
	if err != nil {
		return []Post{}, err
	}
	if len(posts) == 0 {
		return []Post{}, ErrNotFound
	}

	return posts, nil
}

func ReadPostsBySubTime(ctx context.Context, sub string, offset, limit int) ([]Post, error) {
	q := datastore.NewQuery(KindPost).Filter("Sub = ", sub).Order("-Date").Offset(offset).Limit(limit)

	posts := []Post{}
	_, err := ds.GetAll(ctx, q, &posts)
	if err != nil {
		return []Post{}, err
	}
	if len(posts) == 0 {
		return []Post{}, ErrNotFound
	}

	return posts, nil
}

func updateLikePost(ctx context.Context, postID string, update int64) error {
	tx, err := ds.NewTransaction(ctx)
	if err != nil {
		return err
	}

	key := datastore.NameKey(KindPost, postID, nil)

	p := &Post{}
	if err := tx.Get(key, p); err != nil {
		tx.Rollback()
		return err
	}
	p.Post.Likes += update
	if _, err := tx.Put(key, p); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
