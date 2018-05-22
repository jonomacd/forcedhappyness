package handler

import (
	"context"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

type SettingsHandler struct {
	ss sessions.Store
}

func NewSettingsHandler(ss sessions.Store) *SettingsHandler {
	return &SettingsHandler{
		ss: ss,
	}
}

func (h *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.post(w, r)
	case "GET":
		h.get(w, r)
	}
}

func (h *SettingsHandler) post(w http.ResponseWriter, r *http.Request) {
	userID, hasSession := getUserID(w, r, h.ss)
	if !hasSession {
		redirectLogin(w, r)
		return
	}
	r.ParseForm()
	name := r.Form.Get("name")
	username := r.Form.Get("username")
	details := r.Form.Get("details")
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	confirmpassword := r.Form.Get("confirmpassword")

	u := domain.User{
		ID:       userID,
		Email:    email,
		Name:     name,
		Username: username,
		Details:  details,
	}

	if password != confirmpassword {
		tmpl.GetTemplate("settings").Execute(w, domain.BasePage{
			ErrorToast:  "Whoops, those passwords do not match",
			HasSession:  hasSession,
			SessionUser: u,
		})
		return
	}

	if password != "" && len(password) < 4 {
		tmpl.GetTemplate("settings").Execute(w, domain.BasePage{
			ErrorToast:  "Come on... You know that password is too short.",
			HasSession:  hasSession,
			SessionUser: u,
		})
		return
	}

	du := &dao.User{
		User: u,
	}

	if password != "" {
		bb, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("cannot generate password hash %v", err)
			renderError(w, "Whoops, There was a problem trying create user", false)
			return
		}

		du.PasswordHash = bb
	}

	updated, err := dao.UpdateUser(context.Background(), du)
	if err != nil {

		if err == dao.ErrEmailExists || err == dao.ErrUsernameExists {
			tmpl.GetTemplate("settings").Execute(w, domain.BasePage{
				ErrorToast:  "Drat, " + err.Error(),
				HasSession:  hasSession,
				SessionUser: u,
			})
			return
		}

		log.Printf("Bad user: %v", err)
		renderError(w, "Whoops, There was a problem trying to build this page", false)
		return
	}

	http.Redirect(w, r, "/user/"+updated.Username, http.StatusSeeOther)
}

func (h *SettingsHandler) get(w http.ResponseWriter, r *http.Request) {
	userID, ses := getUserID(w, r, h.ss)
	user, err := dao.ReadUserByID(context.Background(), userID)
	if err != nil {
		log.Printf("Cannot read user %v", err)
		renderError(w, "Unable to get user details... soz", ses)
		return
	}
	err = tmpl.GetTemplate("settings").Execute(w, domain.BasePage{
		HasSession:  ses,
		SessionUser: user.User,
	})
	if err != nil {
		log.Printf("Template failed: %v", err)
		renderError(w, "Whoops, There was a problem trying to build this page", ses)
		return
	}
}
