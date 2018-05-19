package handler

import (
	"context"
	"log"
	"net/http"
	"sort"

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
	pg, err := getHomePosts(w, r, ss, sub, "")
	if err != nil {
		log.Printf("Could not get home posts", err)
		return
	}
	err = tmpl.GetTemplate("home").Execute(w, pg)
	if err != nil {
		log.Printf("Template failed: %v", err)
	}
}

func getHomePosts(w http.ResponseWriter, r *http.Request, ss sessions.Store, sub, cursor string) (domain.PageData, error) {
	userID, hasSession := getUserID(w, r, ss)
	if !hasSession {
		// No session means that we will just give them all
		sub = "all"
	}
	ctx := context.Background()

	var err error
	var posts []dao.Post

	if sub == "all" {
		posts, cursor, err = dao.ReadPostsByTime(ctx, cursor, 5)
		if err != nil && err != dao.ErrNotFound {
			log.Printf("Post read failed: %v", err)
		}
	} else if sub != "" {
		posts, cursor, err = dao.ReadPostsBySubTime(ctx, sub, cursor, 100)
		if err != nil && err != dao.ErrNotFound {
			log.Printf("Post read failed: %v", err)
		}
	} else {
		u, err := dao.ReadUserByID(ctx, userID)
		if err != nil {
			log.Printf("could not read user: %v", err)
			return domain.PageData{}, err
		}

		posts, _, err = dao.ReadPostsByUsers(ctx, u.Follows, 20)
		if err != nil && err != dao.ErrNotFound {
			log.Printf("could not read follows: %v", err)
			return domain.PageData{}, err
		}

		mentionPosts, _, err := dao.SearchPostByMention(ctx, u.ID, "", 20)
		if err != nil && err != dao.ErrNotFound {
			log.Printf("could not mention posts: %v", err)
			return domain.PageData{}, err
		}

		posts = append(posts, mentionPosts...)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Date.After(posts[j].Date)
		})
		if len(posts) == 0 {
			renderHome(w, r, ss, "all")
			return domain.PageData{}, nil
		}
	}
	pg := domain.PageData{
		Posts: []domain.PostWithUser{},
		BasePage: domain.BasePage{
			HasSession: hasSession,
		},
		Sub:    sub,
		Cursor: cursor,
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
		post.Post.Text = linkMentionsAndHashtags(post.Post.Text, post.Post.MentionsUsername, post.Post.Hashtags)
		pg.Posts = append(pg.Posts, domain.PostWithUser{
			Post:     post.Post,
			User:     user.User,
			HasLiked: hasLiked,
		})

	}

	return pg, nil
}
