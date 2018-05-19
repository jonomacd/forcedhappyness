package handler

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

type PagingPostHandler struct {
	ss sessions.Store
}

func NewPagingPostHandler(ss sessions.Store) *PagingPostHandler {
	return &PagingPostHandler{
		ss: ss,
	}
}

func (h *PagingPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("loc")
	pathsplit := strings.Split(strings.Trim(path, "/"), "/")
	pathlen := len(pathsplit)

	cursor := r.URL.Query().Get("cursor")

	var err error
	pg := domain.PageData{}
	if pathlen == 2 && pathsplit[0] == "u" {
		pg, err = getHomePosts(w, r, h.ss, pathsplit[1], cursor)
		if err != nil {
			log.Printf("paging failed: %v", err)
			return
		}
	} else if pathlen == 2 && pathsplit[0] == "user" {
		username := pathsplit[1]
		u, err := dao.ReadUserByUsername(context.Background(), username)
		if err != nil {
			log.Printf("error reading userID %s: %v", username, err)
			return
		}
		posts, cursor, err := dao.ReadPostsByUser(context.Background(), u.ID, cursor, 5)

		userID, hasSession := getUserID(w, r, h.ss)
		pg = domain.PageData{
			BasePage: domain.BasePage{
				HasSession: hasSession,
			},
			Cursor: cursor,
		}
		for _, post := range posts {

			hasLiked := false
			if userID != "" {
				_, err := dao.ReadLike(context.Background(), userID, post.ID)
				hasLiked = err == nil
			}

			pg.Posts = append(pg.Posts, domain.PostWithUser{
				Post:     post.Post,
				User:     u.User,
				HasLiked: hasLiked,
			})
		}

	}

	err = tmpl.GetTemplate("postonly").ExecuteTemplate(w, "postsonly", pg)
	if err != nil {
		log.Printf("Template failed: %v", err)
	}

}
