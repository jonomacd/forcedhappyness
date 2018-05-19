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

type LoginHandler struct {
	ss sessions.Store
}

func NewLoginHandler(ss sessions.Store) *LoginHandler {
	return &LoginHandler{
		ss: ss,
	}
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.post(w, r)
	case "GET":
		h.get(w, r)
	}
}

func (h *LoginHandler) post(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	u, err := dao.ReadUserByEmail(context.Background(), email)
	if err != nil {
		if err == dao.ErrNotFound {
			err := tmpl.GetTemplate("login").Execute(w, &loginTemplate{
				BasePage: &domain.BasePage{
					ErrorToast: "Unable to read user",
				},
			})
			if err != nil {
				log.Printf("Template failed: %v", err)
				renderError(w, "Whoops, There was a problem trying to build this page", false)
				return
			}
			return
		}
		log.Printf("Couldn't read user %v", err)
		renderError(w, "Whoops, There was a problem trying to build this page", false)
		return
	}

	redirect := r.Form.Get("redirect")
	if redirect == "" {
		redirect = "/"
	}

	err = bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password))
	if err != nil {
		err := tmpl.GetTemplate("login").Execute(w, &loginTemplate{
			BasePage: &domain.BasePage{
				ErrorToast: "Unable to read user",
			},
		})
		if err != nil {
			log.Printf("Template failed: %v", err)
			renderError(w, "Whoops, There was a problem trying to build this page", false)
			return
		}

		return
	}

	err = newSession(w, r, h.ss, u.ID)
	if err != nil {
		log.Printf("Bad session: %v", err)
		renderError(w, "Whoops, There was a problem trying to build this page", false)
		return
	}

	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

type loginTemplate struct {
	Redirect string
	*domain.BasePage
}

func (h *LoginHandler) get(w http.ResponseWriter, r *http.Request) {

	redirect := r.URL.Query().Get("redirect")

	err := tmpl.GetTemplate("login").Execute(w, &loginTemplate{
		BasePage: &domain.BasePage{},
		Redirect: redirect,
	})
	if err != nil {
		log.Printf("Template failed: %v", err)
	}
}
