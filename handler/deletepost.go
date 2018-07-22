package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/codemodus/parth"
	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/sentiment"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

type DeletePostHandler struct {
	ss sessions.Store
}

func NewDeletePostHandler(ss sessions.Store) *DeletePostHandler {
	return &DeletePostHandler{
		ss: ss,
	}
}

func (h *DeletePostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.post(w, r)
	case "GET":
		h.get(w, r)
	}
}

func renderModeration(w http.ResponseWriter, r *http.Request, ss sessions.Store) {
	userID, hasSession := getUserID(w, r, ss)

	if !hasSession {
		redirectLogin(w, r)
		return
	}
	ctx := context.Background()
	u, err := dao.ReadUserByID(ctx, userID)
	if err != nil {
		renderError(w, "Whoops, can't seem to find your user.", hasSession)
		return
	}

	postID, err := parth.SubSegToString(r.URL.Path, "moderation")
	if err != nil {
		log.Printf("Can't find postid segment %v: %v", postID, err)
		renderError(w, "Whoops, can't seem to find that post.", hasSession)
		return
	}
	p, err := dao.ReadPostByID(ctx, postID)
	if err != nil {
		log.Printf("Can't find post %v: %v", postID, err)
		renderError(w, "Whoops, can't seem to find that post.", hasSession)
		return
	}

	err = validatePostModeration(ctx, u.User, p.Post)
	if err != nil {
		renderError(w, "Looks like you aren't allowed to moderate this post", hasSession)
		return
	}

	postUser, err := dao.ReadUserByID(ctx, p.UserID)
	if err != nil {
		renderError(w, "Whoops, couldn't seem to find the posts user.", hasSession)
		return
	}

	err = tmpl.GetTemplate("moderate").Execute(w, &domain.ModerationPage{
		PostWithUser: domain.PostWithUser{
			Post: p.Post,
			User: postUser.User,
		},
		Redirect: r.URL.Query().Get("redirect"),
		BasePage: &domain.BasePage{
			HasSession: hasSession,
		},
	})
	if err != nil {
		log.Printf("Template failed: %v", err)
	}
}

func (h *DeletePostHandler) get(w http.ResponseWriter, r *http.Request) {
	renderModeration(w, r, h.ss)
}

func (h *DeletePostHandler) post(w http.ResponseWriter, r *http.Request) {
	userID, hasSession := getUserID(w, r, h.ss)

	if !hasSession {
		redirectLogin(w, r)
		return
	}
	ctx := context.Background()
	r.ParseForm()
	text := r.Form.Get("message")
	if len(text) > 500 {
		log.Printf("Too many characters in message %s", text)
		return
	}
	postID := r.Form.Get("post-id")

	url := r.URL.Query().Get("redirect")
	if url == "" {
		url = "/"
	}

	u, err := dao.ReadUserByID(ctx, userID)
	if err != nil {
		log.Printf("unable to read user %s: %v", userID, err)
		return
	}

	p, err := dao.ReadPostByID(ctx, postID)
	if err != nil {
		log.Printf("unable to read post %s: %v", postID, err)
		return
	}

	if err := validatePostModeration(ctx, u.User, p.Post); err != nil {
		renderError(w, "Sorry, you can't delete this post", hasSession)
		return
	}

	if text != "" {
		postSentiment, perspective, err := sentiment.GetNLP(ctx, text)
		if err != nil {
			log.Printf("unable to read sentiment %s: %v", userID, err)
			return
		}

		allowed, _, err := sentiment.CheckPost(ctx, u.User, postSentiment, perspective)
		if err != nil {
			log.Printf("unable check post sentiment %s: %v", userID, err)
			return
		}

		if !allowed {
			renderError(w, "Come on, you are moderating. You are supposed to be the good guy. Be constructive in this message. This is too mean.", hasSession)
			return
		}
	}

	err = dao.SoftDeletePost(ctx, postID, text)
	if err != nil {
		log.Printf("Error soft deleting post %s: %v", postID, err)
		renderError(w, "Whoops, something went wrong deleting that post", hasSession)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

// You have moderation powers on posts if:
// * You wrote the post
// * The post is in your reply chain AND that post hasn't been deleted
// * You are an owner or moderator of the sub the post was posted in
func validatePostModeration(ctx context.Context, user domain.User, post domain.Post) error {
	var err error
	if post.UserID == user.ID {
		// You can delete your own post
		return nil
	}

	if post.IsReply && post.TopParent != post.ID {
		// A parent can delete any child in a reply chain
		var parent dao.Post
		parentID := post.Parent
		for {
			if parent.ID == parentID {
				// In this case, we have read out a post that has the same parent as the previous read.
				// Something has gone wrong in this reply chain or we are at the top.
				break
			}

			parent, err = dao.ReadPostByID(ctx, parentID)
			if err != nil {
				log.Printf("Could not read parent post %s: %v", parentID, err)
				return fmt.Errorf("Unauthorized")
			}

			if parent.UserID == user.ID && !parent.Deleted {
				// If you have been deleted you no longer get moderation rights to the comments.
				// Otherwise, you have access
				return nil
			}

			if parent.ID == post.TopParent {
				// we have reached and checked the top post. No dice.
				break
			}

			parentID = parent.Parent
		}
	}

	if post.Sub != "" {
		sub, err := dao.ReadSub(ctx, post.Sub)
		if err != nil {
			log.Printf("Could not read sub %s: %v", post.Sub, err)
			return fmt.Errorf("Unauthorized")
		}

		for _, owner := range sub.Owners {
			if owner == user.ID {
				// Sub owners can delete posts
				return nil
			}
		}
		for _, mod := range sub.Moderators {
			if mod == user.ID {
				// Sub mods can delete posts
				return nil
			}
		}
	}

	return fmt.Errorf("Unauthorized")
}
