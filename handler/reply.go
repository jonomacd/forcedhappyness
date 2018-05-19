package handler

import (
	"log"
	"net/http"

	"github.com/codemodus/parth"
	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

type ReplyHandler struct {
	ss sessions.Store
}

func NewReplyHandler(ss sessions.Store) *ReplyHandler {
	return &ReplyHandler{
		ss: ss,
	}
}

func (h *ReplyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.get(w, r)
	}
}

func renderReply(w http.ResponseWriter, r *http.Request, ss sessions.Store) {
	_, hasSession := getUserID(w, r, ss)

	if !hasSession {
		redirectLogin(w, r)
		return
	}
	parent, err := parth.SubSegToString(r.URL.Path, "reply")

	if parent == "" || err != nil {
		log.Printf("no parent to reply to")
		return
	}

	err = tmpl.GetTemplate("submit").Execute(w, &domain.Submit{
		ReplyTo: parent,
		BasePage: domain.BasePage{
			HasSession: hasSession,
		},
	})
	if err != nil {
		log.Printf("Template failed: %v", err)
	}
}

func (h *ReplyHandler) get(w http.ResponseWriter, r *http.Request) {
	renderReply(w, r, h.ss)
}
