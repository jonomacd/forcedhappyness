package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/codemodus/parth"
	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

type RottenPostsHandler struct {
	ss sessions.Store
}

func NewRottenPostsHandler(ss sessions.Store) *RottenPostsHandler {
	return &RottenPostsHandler{
		ss: ss,
	}
}

func (h *RottenPostsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	userID, hasSession := getUserID(w, r, h.ss)

	ctx := context.Background()

	cursor := r.URL.Query().Get("cursor")
	next := ""

	username, err := parth.SubSegToString(r.URL.Path, "shame")
	if err != nil {
		log.Printf("error parsing userID %v", err)
		renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
		return
	}

	if username == "me" {
		sessionUser, err := dao.ReadUserByID(ctx, userID)
		if err != nil {
			log.Printf("error reading user %v", err)
			renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
			return
		}
		http.Redirect(w, r, "/shame/"+sessionUser.Username, http.StatusSeeOther)
		return
	}

	u, err := dao.ReadUserByUsername(context.Background(), username)
	if err != nil {

		if err == dao.ErrNotFound {
			renderError(w, "Hmm... this user doesn't appear to exist...", hasSession)
			return
		}
		log.Printf("error reading userID %s: %v", username, err)
		renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
		return
	}

	rps, next, err := dao.ReadRottenPost(ctx, u.User.ID, cursor)
	if err != nil {
		log.Printf("error reading rotten posts %s: %v", username, err)
		renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
		return
	}
	rpwus := []domain.RottenPostWithUser{}
	for _, rp := range rps {

		rpwu := domain.RottenPostWithUser{
			Post: rp.RottenPost,
			User: u.User,
		}
		rpwus = append(rpwus, rpwu)
	}

	pg := domain.RottenPostPageData{
		BasePage: &domain.BasePage{
			HasSession: hasSession,
			Next:       next,
			Previous:   cursor,
		},
		RottenPosts: rpwus,
	}

	err = tmpl.GetTemplate("rottenposts").Execute(w, pg)
	if err != nil {
		log.Printf("Template failed: %v", err)
		renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
		return
	}

}
