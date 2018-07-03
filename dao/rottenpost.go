package dao

import (
	"context"

	"cloud.google.com/go/datastore"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"google.golang.org/api/iterator"
)

var (
	KindRottenPost = "rottenpost"
)

type RottenPost struct {
	domain.RottenPost `datastore:",flatten"`
	K                 *datastore.Key `datastore:"__key__"`
}

func CreateRottenPost(ctx context.Context, p RottenPost) (string, error) {
	if p.ID == "" {
		p.ID = newToken("rp_", 15)
	}

	key := datastore.NameKey(KindRottenPost, p.ID, nil)
	k, err := ds.Put(ctx, key, &p)
	if err != nil {
		return "", err
	}

	return k.Name, err
}

func ReadRottenPost(ctx context.Context, userId, cursor string) ([]RottenPost, string, error) {
	c, err := datastore.DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	q := datastore.NewQuery(KindRottenPost).Filter("UserID = ", userId).Order("-Date").Start(c).Limit(20)

	posts := []RottenPost{}
	it := ds.Run(ctx, q)
	for {
		p := RottenPost{}
		_, err := it.Next(&p)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return []RottenPost{}, "", err
		}

		c, err = it.Cursor()
		if err != nil {
			return []RottenPost{}, "", err
		}
		posts = append(posts, p)
	}

	return posts, c.String(), nil
}
