package handler

import (
	"log"
	"net/http"
	"strings"

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
	log.Printf(r.URL.Path)
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	sub := ""
	if len(parts) == 2 {
		if parts[0] == "u" {
			sub = parts[1]
		}
		renderHome(w, r, h.ss, sub)
	} else if len(parts) == 3 {
		if parts[2] == "post" {
			sub = parts[1]
		}
		renderPost(w, r, h.ss, sub)
	}

}
