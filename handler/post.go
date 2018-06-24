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
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

type PostHandler struct {
	ss sessions.Store
}

func NewPostHandler(ss sessions.Store) *PostHandler {
	return &PostHandler{
		ss: ss,
	}
}

func (h *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	renderPost(w, r, h.ss)

}

func renderPost(w http.ResponseWriter, r *http.Request, ss sessions.Store) {
	userID, hasSession := getUserID(w, r, ss)
	var err error

	postID, err := parth.SubSegToString(r.URL.Path, "post")
	if err != nil {
		renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
		return
	}

	posts, err := dao.ReadPostAndRepliesByID(context.Background(), postID)
	if err != nil {
		log.Printf("Cannot read post %v: %v", postID, err)
		renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
		return
	}
	pg := domain.CommentData{
		BasePage: &domain.BasePage{
			HasSession: hasSession,
		},
	}

	pmap := make(map[string][]domain.PostWithUser, len(posts))
	var topPost domain.PostWithUser

	pwu, err := augmentPosts(context.Background(), userID, posts)
	for _, pu := range pwu {

		pu.Highlight = pu.Post.ID == postID

		if pu.Post.TopParent == pu.Post.ID {
			topPost = pu
			continue
		}
		pmap[pu.Post.Parent] = append(pmap[pu.Post.Parent], pu)
	}

	pg.Post = domain.PostWithComments{
		PostWithUser: topPost,
	}

	populateComments(pmap, &pg.Post)

	err = tmpl.GetTemplate("comments").Execute(w, pg)
	if err != nil {
		log.Printf("Template failed: %v", err)
		renderError(w, "Whoops, There was a problem trying to build this page", hasSession)
		return
	}
}

func populateComments(parentMap map[string][]domain.PostWithUser, pwc *domain.PostWithComments) {
	children, ok := parentMap[pwc.Post.ID]
	if !ok {
		return
	}

	for _, child := range children {
		cpwc := domain.PostWithComments{
			PostWithUser: child,
		}
		populateComments(parentMap, &cpwc)
		pwc.Comments = append(pwc.Comments, cpwc)

	}
}

// TODO: Move to function on post object
func linkMentionsAndHashtags(htmlText template.HTML, mentionsUsername, hashtags []string) template.HTML {

	text := string(htmlText)
	for _, tag := range mentionsUsername {
		text = strings.Replace(text, "@"+tag, "<a class='mention-link' href='/user/"+tag+"'>@"+tag+"</a>", -1)
	}

	for _, tag := range hashtags {
		text = strings.Replace(text, "#"+tag, "<a class='hashtag-link' href='/search?tag="+tag+"'>#"+tag+"</a>", -1)
	}

	return template.HTML(text)
}
