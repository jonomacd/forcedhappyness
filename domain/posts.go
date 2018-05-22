package domain

import (
	"fmt"
	"html/template"
	"time"

	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

type Post struct {
	ID               string
	Text             template.HTML
	Date             time.Time
	UserID           string
	Sub              string
	Reputation       float64
	Likes            int64
	ReplyCount       int64
	IsReply          bool
	TopParent        string
	Parent           string
	Mentions         []string
	MentionsUsername []string
	Hashtags         []string

	// NLP fields
	Analysis *languagepb.AnnotateTextResponse `datastore:"-"`
}

func (p Post) ReplyTo() string {
	return p.ID
}

type PostWithUser struct {
	Post      Post
	User      User
	HasLiked  bool
	Highlight bool
}

type PostWithComments struct {
	PostWithUser
	Comments []PostWithComments
}

func (p Post) FormattedDate() string {
	since := time.Since(p.Date)

	if since < time.Minute*5 {
		return "Just Now"
	}

	if since < time.Hour {
		return fmt.Sprintf("%vm", int(since.Minutes()))
	}

	if since < time.Hour*24 {
		return fmt.Sprintf("%vh", int(since.Hours()))
	}

	if since < time.Hour*24*30*365 {
		return p.Date.Format("Jan 2")
	}

	return p.Date.Format("Jan 2 2006")
}

func (p Post) HappyPercent() int {
	percent := 0
	if p.Analysis != nil && p.Analysis.DocumentSentiment != nil {
		score := p.Analysis.DocumentSentiment.Score
		if score > 0 {
			percent = int(score * 100)
		}
	}

	return percent
}

func (p Post) AngryPercent() int {
	percent := 0
	if p.Analysis != nil && p.Analysis.DocumentSentiment != nil {
		score := p.Analysis.DocumentSentiment.Score
		if score < 0 {
			percent = 100 - int(score*-100)
		}
	}

	return percent
}

type PageData struct {
	*BasePage
	Posts   []PostWithUser
	ReplyTo string
	Sub     string
}

type CommentData struct {
	*BasePage
	Post PostWithComments
}
