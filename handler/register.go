package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

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
	r.ParseForm()
	name := r.Form.Get("name")
	username := r.Form.Get("username")
	details := r.Form.Get("details")
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	confirmpassword := r.Form.Get("confirmpassword")

	if password != confirmpassword {
		log.Printf("non matching passwords")
		return
	}

	u := domain.User{
		Email:        email,
		Name:         name,
		Username:     username,
		Details:      details,
		RegisterDate: time.Now(),
	}

	bb, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("cannot generate password hash %v", err)
	}
	du := &dao.User{
		User:         u,
		PasswordHash: bb,
	}
	err = dao.CreateUser(context.Background(), du)
	if err != nil {
		log.Printf("Bad user: %v", err)
		return
	}

	err = newSession(w, r, h.ss, du.ID)
	if err != nil {
		log.Printf("Bad session: %v", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *RegisterHandler) get(w http.ResponseWriter, r *http.Request) {
	_, ses := getUserID(w, r, h.ss)
	err := tmpl.GetTemplate("register").Execute(w, domain.BasePage{
		HasSession: ses,
	})
	if err != nil {
		log.Printf("Template failed: %v", err)
	}
}
