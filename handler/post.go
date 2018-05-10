package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

type PostHandler struct {
	ss sessions.Store
}

func NewPostHandler(ss sessions.Store) *PostHandler {
	return &PostHandler{
		ss: ss,
	}
}

func (h *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.post(w, r)
	case "GET":
		h.get(w, r)
	}
}

type subData struct {
	Sub string
	domain.BasePage
}

func renderPost(w http.ResponseWriter, r *http.Request, ss sessions.Store, sub string) {
	_, hasSession := getUserID(w, r, ss)

	if !hasSession {
		redirectLogin(w, r)
		return
	}

	err := tmpl.GetTemplate("post").Execute(w, &subData{
		Sub: sub,
		BasePage: domain.BasePage{
			HasSession: hasSession,
		},
	})
	if err != nil {
		log.Printf("Template failed: %v", err)
	}
}

func (h *PostHandler) get(w http.ResponseWriter, r *http.Request) {
	renderPost(w, r, h.ss, "")
}

func (h *PostHandler) post(w http.ResponseWriter, r *http.Request) {
	userID, hasSession := getUserID(w, r, h.ss)

	if !hasSession {
		redirectLogin(w, r)
		return
	}

	r.ParseForm()
	sub := r.Form.Get("sub")
	err := dao.CreatePost(context.Background(), dao.Post{
		Post: domain.Post{
			Date:   time.Now(),
			Text:   r.Form.Get("message"),
			UserID: userID,
			Sub:    sub,
		},
	})
	if err != nil {
		log.Printf("Error writing post: %v", err)
		return
	}
	url := "/"
	if sub != "" {
		url += "u/" + sub + "/"
	}
	http.Redirect(w, r, url, 301)
}
