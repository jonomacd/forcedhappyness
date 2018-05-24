package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

type SearchHandler struct {
	ss sessions.Store
}

func NewSearchHandler(ss sessions.Store) *SearchHandler {

	return &SearchHandler{
		ss: ss,
	}
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()
	userID, hasSession := getUserID(w, r, h.ss)

	tag := r.URL.Query().Get("value")
	params := map[string]string{}
	if tag == "" {
		tag = r.URL.Query().Get("tag")
		params["tag"] = tag
	} else {
		params["value"] = tag
	}

	log.Printf("Searching: %v", tag)

	cursor := r.URL.Query().Get("cursor")
	next := ""
	posts, next, err := dao.ReadPostByHashtag(context.Background(), tag, cursor, 20)

	pg := domain.PageData{
		BasePage: &domain.BasePage{
			HasSession:  hasSession,
			Next:        next,
			Previous:    cursor,
			QueryParams: params,
		},
	}

	pwu, err := augmentPosts(ctx, userID, posts)
	if err != nil {
		log.Printf("user read failed: %v", err)
		renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
		return
	}
	pg.Posts = pwu

	if len(pg.Posts) < 20 {
		pg.Next = ""
	}

	err = tmpl.GetTemplate("home").Execute(w, pg)
	if err != nil {
		renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
		log.Printf("Template failed: %v", err)
	}

}
