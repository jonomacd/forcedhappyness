package dao

import (
	"context"
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/jonomacd/forcedhappyness/site/domain"
)

const (
	KindUser = "user"
)

type User struct {
	domain.User  `datastore:",flatten"`
	PasswordHash []byte

	K *datastore.Key `datastore:"__key__"`
}

func CreateUser(ctx context.Context, u *User) error {
	if u.ID == "" {
		u.ID = newToken("u_", 15)
	}

	key := datastore.NameKey(KindUser, u.ID, nil)
	_, err := ds.Put(ctx, key, u)
	return err
}

func ReadUserByID(ctx context.Context, id string) (User, error) {
	key := datastore.NameKey(KindUser, id, nil)
	u := User{}
	err := ds.Get(ctx, key, &u)
	return u, err
}

func ReadUserByEmail(ctx context.Context, email string) (User, error) {
	q := datastore.NewQuery(KindUser).Filter("Email = ", email)

	users := []User{}
	_, err := ds.GetAll(ctx, q, &users)
	if err != nil {
		return User{}, err
	}
	if len(users) == 0 {
		return User{}, ErrNotFound
	}
	if len(users) > 1 {
		return User{}, fmt.Errorf("Multiple users with same email address")
	}

	return users[0], nil
}
