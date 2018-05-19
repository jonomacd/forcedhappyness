package handler

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gernest/mention"
	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

type SubmitHandler struct {
	ss sessions.Store
}

func NewSubmitHandler(ss sessions.Store) *SubmitHandler {
	return &SubmitHandler{
		ss: ss,
	}
}

func (h *SubmitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.post(w, r)
	case "GET":
		h.get(w, r)
	}
}

func renderSubmit(w http.ResponseWriter, r *http.Request, ss sessions.Store, sub string) {
	_, hasSession := getUserID(w, r, ss)

	if !hasSession {
		redirectLogin(w, r)
		return
	}

	err := tmpl.GetTemplate("submit").Execute(w, &domain.Submit{
		Sub: sub,
		BasePage: domain.BasePage{
			HasSession: hasSession,
		},
	})
	if err != nil {
		log.Printf("Template failed: %v", err)
	}
}

func (h *SubmitHandler) get(w http.ResponseWriter, r *http.Request) {
	renderSubmit(w, r, h.ss, "")
}

func (h *SubmitHandler) post(w http.ResponseWriter, r *http.Request) {
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
	sub := r.Form.Get("sub")
	parent := r.Form.Get("replyto")

	url := "/"
	if sub != "" {
		url += "u/" + sub + "/"
	}

	topParent := ""
	isReply := parent != ""
	if isReply {
		p, err := dao.ReadPostByID(ctx, parent)
		if err != nil {
			log.Printf("Cannot read parent to reply to: %v", err)
			return
		}
		topParent = p.TopParent
		if topParent == "" {
			topParent = parent
		}
		sub = p.Sub
		url += "post/" + topParent + "#" + parent
	}

	sentiment, err := domain.GetNLP(ctx, text)
	if err != nil {
		return
	}

	mentions := parseMentions(text)
	mentionsID := make([]string, 0, len(mentions))
	mentionsUsername := make([]string, 0, len(mentions))
	for _, mention := range mentions {
		u, err := dao.ReadUserByUsername(ctx, mention)
		if err == dao.ErrNotFound {
			continue
		}
		if err != nil {
			log.Printf("Cannot read mentioned user %s: %v", mention, err)
			return
		}
		mentionsID = append(mentionsID, u.ID)
		mentionsUsername = append(mentionsUsername, mention)

	}

	err = dao.CreatePost(ctx, dao.Post{
		Post: domain.Post{
			Date:             time.Now(),
			Text:             template.HTML(text),
			UserID:           userID,
			Sub:              sub,
			IsReply:          isReply,
			Parent:           parent,
			TopParent:        topParent,
			Analysis:         sentiment,
			Mentions:         mentionsID,
			MentionsUsername: mentionsUsername,
			Hashtags:         parseHashtags(text),
		},
	})
	if err != nil {
		log.Printf("Error writing post: %v", err)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

var (
	terminators = []rune{'.', ',', '!', '?', ';', ':', ']', '}', ')'}
)

func parseMentions(text string) []string {
	return mention.GetTagsAsUniqueStrings('@', text, terminators...)
}

func parseHashtags(text string) []string {
	return mention.GetTagsAsUniqueStrings('#', text, terminators...)
}
