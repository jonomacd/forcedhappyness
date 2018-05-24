package dao

import (
	"context"
	"fmt"
	"html/template"

	"cloud.google.com/go/datastore"
	"github.com/jonomacd/forcedhappyness/site/domain"
)

const (
	KindSub = "sub"
)

type Sub struct {
	domain.Sub `datastore:",flatten"`

	K *datastore.Key `datastore:"__key__"`
}

var (
	ErrSubExists = fmt.Errorf("Sub Exists")
)

func CreateSub(ctx context.Context, s Sub) error {

	if s.Name == "" {
		return fmt.Errorf("Must provide a sub name")
	}

	if s.DisplayName == "" {
		s.DisplayName = s.Name
	}

	s.Claimed = true

	_, err := ReadSub(ctx, s.Name)
	if err == nil {
		return ErrSubExists
	}
	if err != nil && err != ErrNotFound {
		return err
	}
	s.Description = template.HTML(template.HTMLEscapeString(string(s.Description)))
	key := datastore.NameKey(KindSub, s.Name, nil)

	_, err = ds.Put(ctx, key, &s)
	return err

}

func ReadSub(ctx context.Context, sub string) (Sub, error) {
	key := datastore.NameKey(KindSub, sub, nil)
	u := Sub{}
	err := ds.Get(ctx, key, &u)
	return u, err
}

func UpdateSub(ctx context.Context, s Sub) error {

	tx, err := ds.NewTransaction(ctx)
	if err != nil {
		return err
	}

	key := datastore.NameKey(KindSub, s.Name, nil)

	current := &Sub{}
	if err := tx.Get(key, current); err != nil {
		tx.Rollback()
		return err
	}

	if s.DisplayName != "" {
		current.DisplayName = s.DisplayName
	}
	if s.Description != "" {
		current.Description = template.HTML(template.HTMLEscapeString(string(s.Description)))
	}

	if len(s.Owners) > 0 {
		toMap := map[string]bool{}
		for _, currOwn := range current.Owners {
			toMap[currOwn] = true

		}
		for _, sOwn := range s.Owners {
			if toMap[sOwn] {
				current.Owners = append(current.Owners, sOwn)
			}
		}
	}

	if len(s.Moderators) > 0 {
		toMap := map[string]bool{}
		for _, currOwn := range current.Moderators {
			toMap[currOwn] = true

		}
		for _, sOwn := range s.Moderators {
			if toMap[sOwn] {
				current.Moderators = append(current.Moderators, sOwn)
			}
		}
	}

	if _, err := tx.Put(key, current); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func DeleteSub(ctx context.Context, sub string) error {
	key := datastore.NameKey(KindSub, sub, nil)
	err := ds.Delete(ctx, key)
	return err
}
