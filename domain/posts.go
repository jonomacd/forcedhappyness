package domain

import (
	"fmt"
	"html/template"
	"time"

	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

type RottenPost struct {
	ID            string
	Text          string
	Date          time.Time
	UserID        string
	RottenGroupId string
}

type Post struct {
	ID               string
	Text             template.HTML `datastore:",noindex"`
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

	LinkDetails []LinkDetails `datastore:",noindex"`

	// DEPRECATED
	Searchtags []string

	// NLP fields
	Analysis *languagepb.AnnotateTextResponse `datastore:"-"`
}

func (p Post) ReplyTo() string {
	return p.ID
}

type LinkDetails struct {
	Url  string
	MIME string
}

func (ld LinkDetails) IsImage() bool {
	switch ld.MIME {
	case "image/bmp", "image/gif", "image/jpeg", "image/tiff", "image/png":
		return true
	}

	return false
}

func (p Post) ImageLinkDetails() []LinkDetails {
	lds := []LinkDetails{}
	for _, ld := range p.LinkDetails {

		if ld.IsImage() {
			lds = append(lds, ld)
		}
		if len(lds) > 3 {
			break
		}
	}

	return lds
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
	SubData Sub
}

type CommentData struct {
	*BasePage
	Post PostWithComments
}

type RottenPostPageData struct {
	*BasePage
	RottenPosts []RottenPostWithUser
}

type RottenPostWithUser struct {
	Post RottenPost
	User User
}

func (p RottenPost) FormattedDate() string {
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
