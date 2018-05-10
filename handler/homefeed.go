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

type HomeFeedHandler struct {
	ss sessions.Store
}

func NewHomeFeedHandler(ss sessions.Store) *HomeFeedHandler {
	return &HomeFeedHandler{
		ss: ss,
	}
}

func (h *HomeFeedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	renderHome(w, r, h.ss, "")

}

func renderHome(w http.ResponseWriter, r *http.Request, ss sessions.Store, sub string) {
	userID, hasSession := getUserID(w, r, ss)
	var err error
	var posts []dao.Post
	if sub == "" {
		posts, err = dao.ReadPostsByTime(context.Background(), 0, 100)
		if err != nil && err != dao.ErrNotFound {
			log.Printf("Post read failed: %v", err)
		}
	} else if sub != "" {
		posts, err = dao.ReadPostsBySubTime(context.Background(), sub, 0, 100)
		if err != nil && err != dao.ErrNotFound {
			log.Printf("Post read failed: %v", err)
		}
	}
	pg := domain.PageData{
		Posts: []domain.PostWithUser{},
		BasePage: domain.BasePage{
			HasSession: hasSession,
		},
	}
	for _, post := range posts {
		user, err := dao.ReadUserByID(context.Background(), post.UserID)
		if err != nil {
			log.Printf("Post read failed: %v", err)
		}

		hasLiked := false
		if userID != "" {
			_, err := dao.ReadLike(context.Background(), userID, post.ID)
			hasLiked = err == nil
		}

		pg.Posts = append(pg.Posts, domain.PostWithUser{
			Post:     post.Post,
			User:     user.User,
			HasLiked: hasLiked,
		})
	}

	err = tmpl.GetTemplate("home").Execute(w, pg)
	if err != nil {
		log.Printf("Template failed: %v", err)
	}
}
