package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
	"github.com/mvader/detect"
)

type NotificationHandler struct {
	ss             sessions.Store
	ApplicationKey string
}

func NewNotificationHandler(ss sessions.Store, publicPushKey string) *NotificationHandler {
	return &NotificationHandler{
		ss:             ss,
		ApplicationKey: publicPushKey,
	}
}

func (h *NotificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, hasSession := getUserID(w, r, h.ss)

	if !hasSession {
		redirectLogin(w, r)
		return
	}
	ctx := context.Background()
	u, err := dao.ReadUserByID(ctx, userID)
	if err != nil {
		log.Printf("Cannot read user %v: %v", userID, err)
		renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
		return
	}
	r.ParseForm()
	if r.Form.Get("_method") == "put" {
		r.Method = "PUT"
	}
	switch r.Method {
	case "POST":
		h.post(ctx, w, r, u)
	case "GET":
		h.get(ctx, w, r, u)
	case "PUT":
		h.put(ctx, w, r, u)
	case "DELETE":
		h.delete(ctx, w, r, u)

	}
}

func (h NotificationHandler) post(ctx context.Context, w http.ResponseWriter, r *http.Request, u dao.User) {
	notification := parseNotificationForm(r, u)

	if err := dao.CreateNotification(ctx, dao.Notification{Notification: notification}); err != nil {
		log.Printf("Cannot create notificaion %#v: %v", notification, err)
		renderError(w, "Whoops, There was a problem trying to build this page", true)
		return
	}
	redirect := r.Form.Get("redirect")
	if redirect == "" {
		redirect = "/notifications"
	}
	// After create render the notification page
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (h NotificationHandler) get(ctx context.Context, w http.ResponseWriter, r *http.Request, u dao.User) {
	notifications, err := dao.ReadNotifications(ctx, u.User.ID)
	if err != nil {
		log.Printf("Unable to read notifications: %v", err)
		renderError(w, "Whoops, There was a problem trying to build this page", true)
		return
	}

	n := make([]domain.Notification, len(notifications))
	for ii, notif := range notifications {
		n[ii] = notif.Notification
	}

	np := domain.NotificationPage{
		BasePage: &domain.BasePage{
			HasSession:  true,
			SessionUser: u.User,
		},
		Notifications:  n,
		ApplicationKey: h.ApplicationKey,
		Redirect:       r.URL.Query().Get("redirect"),
	}

	err = tmpl.GetTemplate("notifications").Execute(w, np)
	if err != nil {
		log.Printf("Template failed: %v", err)
		renderError(w, "Whoops, There was a problem trying to build this page", true)
	}

}

func (h NotificationHandler) put(ctx context.Context, w http.ResponseWriter, r *http.Request, u dao.User) {
	notification := parseNotificationForm(r, u)

	// TODO: only update what needs updating
	err := dao.UpdateNotification(ctx, notification.ID, nil, notification.Config)
	if err != nil {
		log.Printf("Failed update %v\n", err)
		renderError(w, "Whoops, There was a problem trying to build this page", true)
		return
	}

	redirect := r.Form.Get("redirect")
	if redirect == "" {
		redirect = "/notifications"
	}
	// After create render the notification page
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (h NotificationHandler) delete(ctx context.Context, w http.ResponseWriter, r *http.Request, u dao.User) {
	r.ParseForm()
	id := r.Form.Get("notificationID")
	if id != "" {
		err := dao.DeleteNotification(ctx, id)
		if err != nil {
			renderError(w, "Whoops, There was a problem trying to build this page", true)
			return
		}
	}

	// After delete render the notification page
	redirect := r.Form.Get("redirect")
	if redirect == "" {
		redirect = "/notifications"
	}
	// After create render the notification page
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func parseNotificationForm(r *http.Request, u dao.User) domain.Notification {
	form := r.Form
	sub := form.Get("hidden-sub")
	var wsub *webpush.Subscription
	if sub != "" {
		wsub = &webpush.Subscription{}
		err := json.Unmarshal([]byte(sub), wsub)
		if err != nil {
			log.Printf("Unable to unmarshal subscription %s: %v", sub, err)
		}
	}

	return domain.Notification{
		ID:     form.Get("id"),
		Name:   detect.Platform(r.UserAgent()).String(),
		UserID: u.User.ID,
		Config: &domain.NotificationConfiguration{
			FollowerPost: mustParseBool(form.Get("followerpost")),
			FollowerGet:  mustParseBool(form.Get("followerget")),
			Likes:        mustParseBool(form.Get("likes")),
			Mentions:     mustParseBool(form.Get("mentions")),
			Replies:      mustParseBool(form.Get("replies")),
		},
		Subscription: wsub,
	}
}

func mustParseBool(value string) bool {
	r, _ := strconv.ParseBool(value)
	if value == "on" {
		r = true
	}
	return r
}
