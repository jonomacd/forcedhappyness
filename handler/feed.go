package handler

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/dyatlov/go-oembed/oembed"

	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/embedder"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
	"mvdan.cc/xurls"
)

var (
	urlMatcher = xurls.Relaxed()
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
	log.Printf("Got request to home handler")
	renderHome(w, r, h.ss, "")

}

func renderHome(w http.ResponseWriter, r *http.Request, ss sessions.Store, sub string) {

	userID, hasSession := getUserID(w, r, ss)
	if !hasSession {
		// No session means that we will just give them all
		sub = "all"
	}
	ctx := context.Background()

	cursor := r.URL.Query().Get("cursor")
	next := ""
	var err error
	var posts []dao.Post
	var subData dao.Sub
	if sub == "all" {
		posts, next, err = dao.ReadPostsByTime(ctx, cursor, 20)
		if err != nil && err != dao.ErrNotFound {
			log.Printf("Post read by time failed: %v", err)
			renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
			return
		}
		sub = ""
	} else if sub != "" {
		posts, next, err = dao.ReadPostsBySubTime(ctx, sub, cursor, 20)
		if err != nil && err != dao.ErrNotFound {
			log.Printf("Post read by sub failed: %v", err)
			renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
			return
		}

		subData, err = dao.ReadSub(ctx, sub)
		if err != nil && err != dao.ErrNotFound {
			log.Printf("Post read by sub failed: %v", err)
			renderError(w, "Whoops, There was a problem trying to build this sub", hasSession)
			return
		}

		subData.Description, _ = augmentWithLinks(subData.Description)

	} else {
		u, err := dao.ReadUserByID(ctx, userID)
		if err != nil {
			log.Printf("could not read user: %v", err)
			renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
			return
		}

		posts, next, err = ReadUserMentionAndFollowPosts(ctx, u, cursor)
		if err != nil {
			log.Printf("could not read post,mention,user: %v", err)
			renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
			return
		}
		if len(u.Follows) == 0 {
			http.Redirect(w, r, "/n/all", http.StatusSeeOther)
			return
		}
	}
	pg := domain.PageData{
		BasePage: &domain.BasePage{
			HasSession: hasSession,
			Next:       next,
			Previous:   cursor,
		},
		Sub:     sub,
		SubData: subData.Sub,
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

func ReadUserMentionAndFollowPosts(ctx context.Context, u dao.User, cursor string) ([]dao.Post, string, error) {

	cc, err := dao.ReadCursor(ctx, cursor)
	if err != nil {
		return nil, "", err
	}

	mentionPosts, mentionCursor, err := dao.SearchPostByMention(ctx, u.ID, cc["mentions"], 20)
	if err != nil && err != dao.ErrNotFound {
		log.Printf("could not mention posts: %v", err)
		return nil, "", err
	}

	posts, cursors, err := dao.ReadPostsByUsers(ctx, u.Follows, cc, 20)
	if err != nil && err != dao.ErrNotFound {
		log.Printf("could not read follows: %v", err)
		return nil, "", err
	}

	selfPosts, selfCursor, err := dao.ReadPostsByUser(ctx, u.ID, cc["self"], 20)
	if err != nil && err != dao.ErrNotFound {
		log.Printf("could not read self posts: %v", err)
		return nil, "", err
	}

	cursors["mentions"] = mentionCursor
	cursors["self"] = selfCursor

	cursor, err = dao.CreateCursor(ctx, cursors)
	if err != nil && err != dao.ErrNotFound {
		log.Printf("could not create cursor: %v", err)
		return nil, "", err
	}

	posts = append(posts, mentionPosts...)
	posts = append(posts, selfPosts...)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	return posts, cursor, nil

}

func augmentPosts(ctx context.Context, userID string, posts []dao.Post) ([]domain.PostWithUser, error) {
	pwu := make([]domain.PostWithUser, len(posts))
	u, err := dao.ReadUserByID(ctx, userID)
	if err != nil {
		log.Printf("cannot read user %s: %v", userID, err)
	}
	for ii, post := range posts {
		user, err := dao.ReadUserByID(ctx, posts[ii].UserID)
		if err != nil {
			return nil, err
		}

		hasLiked := false
		if userID != "" {
			_, err := dao.ReadLike(ctx, userID, posts[ii].ID)
			hasLiked = err == nil
		}

		canModerate := false
		if u.ID != "" {
			err = validatePostModeration(ctx, u.User, post.Post)
			canModerate = err == nil
		}

		posts[ii].Post.Text, posts[ii].Post.Embed = augmentWithLinks(posts[ii].Post.Text)
		posts[ii].Post.Text = linkMentionsAndHashtags(posts[ii].Post.Text, posts[ii].Post.MentionsUsername, posts[ii].Post.Hashtags)
		pwu[ii] = domain.PostWithUser{
			Post:        posts[ii].Post,
			User:        user.User,
			HasLiked:    hasLiked,
			CanModerate: canModerate,
		}
	}

	return pwu, nil
}

func augmentWithLinks(text template.HTML) (template.HTML, template.HTML) {

	embed := ""

	linktext := template.HTML(urlMatcher.ReplaceAllStringFunc(string(text), func(in string) string {

		if embed == "" {
			toEmbed := embedder.Oembed.FindItem(in)

			if toEmbed != nil {
				if toEmbed.ProviderName == "Twitter" {
					if !strings.Contains(toEmbed.EndpointURL, "omit_script") {
						toEmbed.EndpointURL = toEmbed.EndpointURL[:len(toEmbed.EndpointURL)-5] + "&hide_thread=true&dnt=true&omit_script=true&url="
						log.Printf("Endpoint URL: %s", toEmbed.EndpointURL)
					}
				}
				i, err := toEmbed.FetchOembed(oembed.Options{

					URL: in,
				})
				if err != nil {
					log.Printf("embed error: %v\n", err)
				} else {
					embed = i.HTML
				}
			}
		}

		linkPrefix := ""
		if !strings.HasPrefix(in, "http") {
			linkPrefix = "//"
		}
		return "<a target='_blank' href='" + linkPrefix + in + "'>" + in + "</a>"
	}))

	return linktext, template.HTML(embed)
}
