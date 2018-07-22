package dao

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/jonomacd/forcedhappyness/site/domain"
)

const (
	KindUser      = "user"
	KindFollowers = "followers"
)

var (
	ErrUsernameExists = fmt.Errorf("Username Exists")
	ErrEmailExists    = fmt.Errorf("Email Exists")
)

type Followers struct {
	FollowedID string
	FollowerID string
	K          *datastore.Key `datastore:"__key__"`
}

type User struct {
	domain.User  `datastore:",flatten"`
	PasswordHash []byte
	SignInMethod string

	K *datastore.Key `datastore:"__key__"`
}

func CreateUser(ctx context.Context, u *User) error {
	if u.ID == "" {
		u.ID = newToken("u_", 15)
	}

	u.Username = strings.ToLower(u.Username)
	u.Email = strings.ToLower(u.Email)

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

func UpdateUser(ctx context.Context, u *User) (*User, error) {
	tx, err := ds.NewTransaction(ctx)
	if err != nil {
		return nil, err
	}

	u.Username = strings.ToLower(u.Username)
	u.Email = strings.ToLower(u.Email)

	key := datastore.NameKey(KindUser, u.User.ID, nil)

	current := &User{}
	if err := tx.Get(key, current); err != nil {
		tx.Rollback()
		return nil, err
	}

	if u.PasswordHash != nil {
		current.PasswordHash = u.PasswordHash
	}
	if u.Name != "" && u.Name != current.Name {
		current.Name = u.Name
	}
	if u.Email != "" && u.Email != current.Email {
		_, err := ReadUserByEmail(ctx, u.Email)
		if err != datastore.ErrNoSuchEntity {
			return nil, ErrEmailExists
		}
		current.Email = u.Email
	}

	if u.Username != "" && u.Username != current.Username {
		_, err = ReadUserByUsername(ctx, u.Username)
		if err != datastore.ErrNoSuchEntity {
			return nil, ErrUsernameExists
		}
		current.Username = u.Username
	}

	if u.Details != "" && u.Details != current.Details {
		current.Details = u.Details
	}

	if u.Avatar != "" && u.Avatar != current.Avatar {
		current.Avatar = u.Avatar
	}

	if _, err := tx.Put(key, current); err != nil {
		tx.Rollback()
		return nil, err
	}
	if _, err = tx.Commit(); err != nil {
		return nil, err
	}
	return current, nil
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

	key = datastore.NameKey(KindFollowers, newToken("f_", 10), nil)
	ds.Put(ctx, key, &Followers{
		FollowedID: followUserID,
		FollowerID: userID,
	})

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

	q := datastore.NewQuery(KindFollowers).Filter("FollowerID = ", userID).Filter("FollowedID = ", followUserID)
	f := []Followers{}
	_, err = ds.GetAll(ctx, q, &f)
	if err != nil {
		return err
	}
	if len(f) > 0 {
		keys := []*datastore.Key{}
		for _, ff := range f {
			keys = append(keys, ff.K)
		}
		err = ds.DeleteMulti(ctx, keys)
		if err != nil {
			return err
		}
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

func updateUserPostedStatistics(ctx context.Context, p Post, userID string) error {
	tx, err := ds.NewTransaction(ctx)
	if err != nil {
		return err
	}

	key := datastore.NameKey(KindUser, userID, nil)

	current := &User{}
	if err := tx.Get(key, current); err != nil {
		tx.Rollback()
		return err
	}

	current.PostedCount++
	current.PostedSentiment += float64(p.Analysis.DocumentSentiment.Score)

	if _, err := tx.Put(key, current); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func UpdateUserPostAttemptStatistics(ctx context.Context, user User, score float32) error {
	tx, err := ds.NewTransaction(ctx)
	if err != nil {
		return err
	}

	key := datastore.NameKey(KindUser, user.ID, nil)

	current := &User{}
	if err := tx.Get(key, current); err != nil {
		tx.Rollback()
		return err
	}

	current.PostAttemptCount++
	current.PostAttemptSentiment += float64(score)

	current.AngryBanExpire = user.AngryBanExpire
	current.AngryBanThreshold = user.AngryBanThreshold
	current.AngryBanCount = user.AngryBanCount

	current.TotalSentimentEMA = user.TotalSentimentEMA

	if _, err := tx.Put(key, current); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func ReadFollowers(ctx context.Context, userID string) ([]string, error) {
	q := datastore.NewQuery(KindFollowers).Filter("FollowedID = ", userID)

	f := []Followers{}
	_, err := ds.GetAll(ctx, q, &f)
	if err != nil {
		return nil, err
	}
	if len(f) == 0 {
		return nil, ErrNotFound
	}
	followers := []string{}

	for _, ff := range f {
		followers = append(followers, ff.FollowerID)
	}

	return followers, nil
}
