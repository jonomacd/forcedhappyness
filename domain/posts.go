package domain

import (
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

func (post Post) FormattedDate() string {
	return post.Date.Format(time.RFC822)
}

type PageData struct {
	BasePage
	Posts   []PostWithUser
	ReplyTo string
	Sub     string
}

type CommentData struct {
	BasePage
	Post PostWithComments
}
