package handler

import (
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
)

const (
	sessionName      = "iwbn-session"
	sessionUserIDKey = "userID"
)

func getUserID(w http.ResponseWriter, r *http.Request, ss sessions.Store) (string, bool) {
	session, err := ss.Get(r, sessionName)
	if err != nil {
		s, _ := ss.New(r, sessionName)
		s.Options.MaxAge = -1
		s.Save(r, w)
		return "", false
	}
	if uidi, ok := session.Values[sessionUserIDKey]; ok {
		return uidi.(string), true
	}
	return "", false
}

func newSession(w http.ResponseWriter, r *http.Request, ss sessions.Store, userID string) error {
	s, err := ss.New(r, sessionName)
	if err != nil {
		return err
	}
	s.Values[sessionUserIDKey] = userID
	return s.Save(r, w)
}

func redirectLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login?redirect="+url.PathEscape(r.URL.Path), http.StatusSeeOther)
}
