package domain

import (
	webpush "github.com/SherClockHolmes/webpush-go"
)

type Notification struct {
	ID           string
	UserID       string
	Name         string                     `datastore:",noindex"`
	Subscription *webpush.Subscription      `datastore:",noindex"`
	Config       *NotificationConfiguration `datastore:",noindex"`
}

type NotificationConfiguration struct {
	Replies      bool
	Likes        bool
	FollowerPost bool
	FollowerGet  bool
	Mentions     bool
	Subs         []string
	Users        []string
}

type NotificationPage struct {
	*BasePage
	Notifications  []Notification
	ApplicationKey string
	Redirect       string
}
