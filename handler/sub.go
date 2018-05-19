package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/codemodus/parth"
	"github.com/gorilla/sessions"
)

type SubHandler struct {
	ss sessions.Store
}

func NewSubHandler(ss sessions.Store) *SubHandler {

	return &SubHandler{
		ss: ss,
	}
}

func (h *SubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if forceTrailingSlash(w, r) {
		return
	}

	pathLen := len(strings.Split(strings.Trim(r.URL.Path, "/"), "/"))
	sub, err := parth.SubSegToString(r.URL.Path, "u")
	if err != nil {
		log.Printf("error getting sub: %v", err)
		return
	}
	switch pathLen {
	case 2:
		renderHome(w, r, h.ss, sub)
	case 3:
		action, err := parth.SegmentToString(r.URL.Path, 2)
		if err != nil {
			log.Printf("error getting action: %v", err)
			return
		}
		if action == "submit" {
			renderSubmit(w, r, h.ss, sub)
			return
		}

	}
}

func forceTrailingSlash(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasSuffix(r.URL.Path, "/") {
		r.URL.Path += "/"
		http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
		return true
	}
	return false
}
