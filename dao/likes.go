package dao

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/jonomacd/forcedhappyness/site/domain"
)

const (
	KindLike = "like"
)

type Like struct {
	domain.Like `datastore:",flatten"`

	K *datastore.Key `datastore:"__key__"`
}

func CreateLike(ctx context.Context, userID, postID string) error {

	l := &Like{
		Like: domain.Like{
			PostID: postID,
			UserID: userID,
			Date:   time.Now(),
		},
	}

	key := datastore.NameKey(KindLike, userID+postID, nil)
	_, err := ds.Put(ctx, key, l)
	if err != nil {
		return err
	}

	return updateLikePost(ctx, postID, 1)
}

func ReadLike(ctx context.Context, userID, postID string) (*Like, error) {
	key := datastore.NameKey(KindLike, userID+postID, nil)
	l := &Like{}
	err := ds.Get(ctx, key, l)
	return l, err
}

func DeleteLike(ctx context.Context, userID, postID string) error {
	key := datastore.NameKey(KindLike, userID+postID, nil)
	err := ds.Delete(ctx, key)
	if err != nil {
		return err
	}

	return updateLikePost(ctx, postID, -1)
}
