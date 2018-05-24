package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/tmpl"

	"github.com/codemodus/parth"
)

type UserHandler struct {
	ss sessions.Store
}

func NewUserHandler(ss sessions.Store) *UserHandler {
	return &UserHandler{
		ss: ss,
	}
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	renderUser(w, r, h.ss)

}

func renderUser(w http.ResponseWriter, r *http.Request, ss sessions.Store) {
	if forceTrailingSlash(w, r) {
		return
	}

	ctx := context.Background()
	userID, hasSession := getUserID(w, r, ss)

	cursor := r.URL.Query().Get("cursor")

	username, err := parth.SubSegToString(r.URL.Path, "user")
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
		http.Redirect(w, r, "/user/"+sessionUser.Username, http.StatusSeeOther)
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

	follows := false
	sessionUser := dao.User{}
	if hasSession {
		sessionUser, err = dao.ReadUserByID(ctx, userID)
		if err != nil {
			log.Printf("Unable to read session user %s: %v", userID, err)
			renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
			return
		}

		for _, id := range sessionUser.Follows {
			if id == u.ID {
				follows = true
				break
			}
		}
	}

	posts, next, err := dao.ReadPostsByUser(ctx, u.ID, cursor, 20)
	pg := domain.UserPage{
		User: u.User,
		BasePage: &domain.BasePage{
			HasSession:  hasSession,
			SessionUser: sessionUser.User,
			Next:        next,
			Previous:    cursor,
		},
		Follows: follows,
	}
	for _, post := range posts {

		hasLiked := false
		if userID != "" {
			_, err := dao.ReadLike(ctx, userID, post.ID)
			hasLiked = err == nil
		}

		pg.Posts = append(pg.Posts, domain.PostWithUser{
			Post:     post.Post,
			User:     u.User,
			HasLiked: hasLiked,
		})
	}

	err = tmpl.GetTemplate("user").Execute(w, pg)
	if err != nil {
		log.Printf("Template failed: %v", err)
		renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
	}
}
