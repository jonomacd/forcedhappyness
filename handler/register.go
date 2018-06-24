package handler

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	oauth2 "google.golang.org/api/oauth2/v2"

	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

type RegisterHandler struct {
	ss sessions.Store
}

func NewRegisterHandler(ss sessions.Store) *RegisterHandler {
	return &RegisterHandler{
		ss: ss,
	}
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.post(w, r)
	case "GET":
		h.get(w, r)
	}
}

func (h *RegisterHandler) post(w http.ResponseWriter, r *http.Request) {

	_, hasSession := getUserID(w, r, h.ss)
	if hasSession {
		renderError(w, "Good news! You already have an account.", hasSession)
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
		Email:        email,
		Name:         name,
		Username:     username,
		Details:      details,
		RegisterDate: time.Now(),
	}

	if password != confirmpassword {
		log.Printf("non matching passwords")
		tmpl.GetTemplate("register").Execute(w, domain.BasePage{
			ErrorToast:  "Whoops, those passwords do not match",
			SessionUser: u,
		})
		return
	}

	if password != "" && len(password) < 5 {
		tmpl.GetTemplate("register").Execute(w, domain.BasePage{
			ErrorToast:  "Come on... You know that password is too short.",
			SessionUser: u,
		})
		return
	}

	bb, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("cannot generate password hash %v", err)
		renderError(w, "Whoops, There was a problem trying create user", false)
		return
	}
	du := &dao.User{
		User:         u,
		PasswordHash: bb,
	}
	err = dao.CreateUser(context.Background(), du)
	if err != nil {

		if err == dao.ErrEmailExists || err == dao.ErrUsernameExists {
			tmpl.GetTemplate("register").Execute(w, domain.BasePage{
				ErrorToast:  "Drat, " + err.Error(),
				SessionUser: u,
			})
			return
		}

		log.Printf("Bad user: %v", err)
		renderError(w, "Whoops, There was a problem trying to build this page", false)
		return
	}

	err = newSession(w, r, h.ss, du.ID)
	if err != nil {
		log.Printf("Bad session: %v", err)
		renderError(w, "Whoops, There was a problem trying to build this page", false)
		return
	}

	http.Redirect(w, r, "/welcome", http.StatusSeeOther)
}

func (h *RegisterHandler) get(w http.ResponseWriter, r *http.Request) {
	_, ses := getUserID(w, r, h.ss)
	if ses {
		renderError(w, "Good news! You already have an account.", ses)
		return
	}

	err := tmpl.GetTemplate("register").Execute(w, domain.BasePage{
		HasSession: ses,
	})
	if err != nil {
		log.Printf("Template failed: %v", err)
		renderError(w, "Whoops, There was a problem trying to build this page", ses)
		return
	}
}

type GoogleRegisterHandler struct {
	ss sessions.Store
}

func NewGoogleRegisterHandler(ss sessions.Store) *GoogleRegisterHandler {
	return &GoogleRegisterHandler{
		ss: ss,
	}
}

type googleIdToken struct {
	IdToken string `json:"id_token"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
}

func (h *GoogleRegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	r.ParseForm()
	git := &googleIdToken{
		Avatar:  r.Form.Get("avatar"),
		IdToken: r.Form.Get("id_token"),
		Name:    r.Form.Get("name"),
	}

	info, err := verifyIdToken(git.IdToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	u, err := dao.ReadUserByEmail(ctx, info.Email)
	if err != nil && err != dao.ErrNotFound {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err == dao.ErrNotFound {
		// User Not Found, So let's create one
		du := &dao.User{
			User: domain.User{
				Email:        info.Email,
				Name:         git.Name,
				Username:     strings.Replace(git.Name, " ", "", -1),
				Avatar:       git.Avatar,
				RegisterDate: time.Now(),
			},
			SignInMethod: "google",
		}

		err = dao.CreateUser(ctx, du)
		if err != nil {
			if err == dao.ErrEmailExists || err == dao.ErrUsernameExists {
				log.Printf("User exists: %v", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			log.Printf("Bad user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = newSession(w, r, h.ss, du.ID)
		if err != nil {
			log.Printf("Bad session new user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/welcome", http.StatusSeeOther)
		return
	} else {
		// User is found. Let's set up a session.
		err = newSession(w, r, h.ss, u.ID)
		if err != nil {
			log.Printf("Bad session existing user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		redirect := r.URL.Query().Get("redirect")
		if redirect == "" {
			redirect = "/"
		}
		http.Redirect(w, r, redirect, http.StatusSeeOther)
		return
	}

}

var httpClient = &http.Client{}

func verifyIdToken(idToken string) (*oauth2.Tokeninfo, error) {
	oauth2Service, err := oauth2.New(httpClient)
	tokenInfoCall := oauth2Service.Tokeninfo()
	tokenInfoCall.IdToken(idToken)
	tokenInfo, err := tokenInfoCall.Do()
	if err != nil {
		return nil, err
	}
	return tokenInfo, nil
}
