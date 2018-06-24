package handler

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/codemodus/parth"
	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
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
	_, hasSession := getUserID(w, r, h.ss)
	sub, err := parth.SubSegToString(r.URL.Path, "n")
	if err != nil {
		log.Printf("error getting sub: %v", err)
		renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
		return
	}

	renderHome(w, r, h.ss, sub)
}

func forceTrailingSlash(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasSuffix(r.URL.Path, "/") {
		r.URL.Path += "/"
		http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
		return true
	}
	return false
}

type SubCRUDHandler struct {
	ss sessions.Store
}

func NewSubCRUDHandler(ss sessions.Store) *SubCRUDHandler {

	return &SubCRUDHandler{
		ss: ss,
	}
}

func (h *SubCRUDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.post(w, r)
	case "PUT":
		h.put(w, r)
	case "DELETE":
		h.delete(w, r)
	}
}

func (h *SubCRUDHandler) post(w http.ResponseWriter, r *http.Request) {
	userID, hasSession := getUserID(w, r, h.ss)
	if !hasSession {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("No session for creating sub")
		return
	}

	ctx := context.Background()
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Unable to parse form for sub %v", err)
		return
	}

	name := r.Form.Get("name")
	description := r.Form.Get("description")
	sub := r.Form.Get("sub")

	err = dao.CreateSub(ctx, dao.Sub{
		Sub: domain.Sub{
			Name:        sub,
			Description: template.HTML(description),
			DisplayName: name,
			Owners:      []string{userID},
		},
	})
	if err != nil {
		if err == dao.ErrSubExists {
			w.WriteHeader(http.StatusConflict)
			log.Printf("Conflict, sub already exists %v", err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Unable to create for sub %v", err)
		return
	}

	http.Redirect(w, r, "/n/"+sub, http.StatusSeeOther)
}

func (h *SubCRUDHandler) put(w http.ResponseWriter, r *http.Request) {
	userID, hasSession := getUserID(w, r, h.ss)
	if !hasSession {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	name := r.Form.Get("name")
	description := r.Form.Get("description")
	sub := r.Form.Get("sub")

	s, err := dao.ReadSub(ctx, sub)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	allowed := false
	for _, owner := range s.Owners {
		if owner == userID {
			allowed = true
			break
		}
	}

	if !allowed {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = dao.UpdateSub(ctx, dao.Sub{
		Sub: domain.Sub{
			Name:        sub,
			Description: template.HTML(description),
			DisplayName: name,
		},
	})
	if err != nil {
		if err == dao.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderHome(w, r, h.ss, sub)
}

func (h *SubCRUDHandler) delete(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	userID, hasSession := getUserID(w, r, h.ss)
	if !hasSession {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	sub := r.URL.Query().Get("sub")
	if sub == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s, err := dao.ReadSub(ctx, sub)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	allowed := false
	for _, owner := range s.Owners {
		if owner == userID {
			allowed = true
			break
		}
	}

	if !allowed {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = dao.DeleteSub(ctx, sub)
	if err != nil {
		if err == dao.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
