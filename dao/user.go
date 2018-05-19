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

var (
	ErrUsernameExists = fmt.Errorf("Username Exists")
	ErrEmailExists    = fmt.Errorf("Email Exists")
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

	_, err := ReadUserByEmail(ctx, u.Email)
	if err != datastore.ErrNoSuchEntity {
		return ErrEmailExists
	}

	_, err = ReadUserByUsername(ctx, u.Username)
	if err != datastore.ErrNoSuchEntity {
		return ErrUsernameExists
	}

	key := datastore.NameKey(KindUser, u.ID, nil)
	_, err = ds.Put(ctx, key, u)
	return err
}

func AddFollowerToUser(ctx context.Context, userID, followUserID string) error {
	tx, err := ds.NewTransaction(ctx)
	if err != nil {
		return err
	}

	key := datastore.NameKey(KindUser, userID, nil)

	p := &User{}
	if err := tx.Get(key, p); err != nil {
		tx.Rollback()
		return err
	}
	p.User.Follows = append(p.User.Follows, followUserID)
	if _, err := tx.Put(key, p); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func DeleteFollowerFromUser(ctx context.Context, userID, followUserID string) error {
	tx, err := ds.NewTransaction(ctx)
	if err != nil {
		return err
	}

	key := datastore.NameKey(KindUser, userID, nil)

	p := &User{}
	if err := tx.Get(key, p); err != nil {
		tx.Rollback()
		return err
	}
	p.User.Follows = remove(p.User.Follows, followUserID)
	if _, err := tx.Put(key, p); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
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

func ReadUserByUsername(ctx context.Context, username string) (User, error) {
	q := datastore.NewQuery(KindUser).Filter("Username = ", username)

	users := []User{}
	_, err := ds.GetAll(ctx, q, &users)
	if err != nil {
		return User{}, err
	}
	if len(users) == 0 {
		return User{}, ErrNotFound
	}
	if len(users) > 1 {
		return User{}, fmt.Errorf("Multiple users with same username")
	}

	return users[0], nil
}
