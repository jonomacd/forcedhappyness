package dao

import (
	"context"
	"fmt"

	"cloud.google.com/go/datastore"
	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/jonomacd/forcedhappyness/site/domain"
)

const (
	KindNotification = "notification"
)

type Notification struct {
	domain.Notification `datastore:",flatten"`

	K *datastore.Key `datastore:"__key__"`
}

func CreateNotification(ctx context.Context, notification Notification) error {
	if notification.UserID == "" {
		return fmt.Errorf("no user specified")
	}
	if notification.Name == "" {
		return fmt.Errorf("no device name")
	}
	if notification.Subscription == nil {
		return fmt.Errorf("no subscription specified")
	}

	notification.ID = newToken("n_", 10)
	key := datastore.NameKey(KindNotification, notification.ID, nil)
	_, err := ds.Put(ctx, key, &notification)
	if err != nil {
		return err
	}

	return nil
}

func ReadNotification(ctx context.Context, id string) (Notification, error) {
	key := datastore.NameKey(KindNotification, id, nil)

	notification := Notification{}
	err := ds.Get(ctx, key, &notification)

	return notification, err
}

func ReadNotifications(ctx context.Context, userID string) ([]Notification, error) {
	q := datastore.NewQuery(KindNotification).Filter("UserID =", userID)

	notifications := []Notification{}
	_, err := ds.GetAll(ctx, q, &notifications)

	return notifications, err
}

func UpdateNotification(ctx context.Context, id string, subscription *webpush.Subscription, config *domain.NotificationConfiguration) error {
	tx, err := ds.NewTransaction(ctx)
	if err != nil {
		return err
	}

	key := datastore.NameKey(KindNotification, id, nil)

	notification := Notification{}
	if err := tx.Get(key, &notification); err != nil {
		tx.Rollback()
		return err
	}

	if subscription != nil {
		notification.Subscription = subscription
	}
	if config != nil {
		notification.Config = config
	}
	if _, err := tx.Put(key, &notification); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func DeleteNotification(ctx context.Context, id string) error {
	key := datastore.NameKey(KindNotification, id, nil)
	err := ds.Delete(ctx, key)
	if err != nil {
		return err
	}

	return nil
}
