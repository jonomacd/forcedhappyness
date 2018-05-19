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

	ss := ""
	hours := int(since.Hours())
	if hours > 0 {
		days := int(hours / 24)
		if days > 0 {
			years := int(days / 365)
			if years > 0 {
				s := ""
				if years > 1 {
					s = "s"
				}
				ss += fmt.Sprintf("%v year%s ", years, s)
			}
			s := ""
			if years > 1 {
				s = "s"
			}
			daysp := days
			if years > 0 {
				daysp %= years * 365
			}
			ss += fmt.Sprintf("%v day%s ", daysp, s)
			if years > 0 {
				return ss + "ago"
			}
		}
		s := ""
		if days > 1 {
			s = "s"
		}
		hoursp := hours
		if days > 0 {
			hoursp %= days * 24
		}
		ss += fmt.Sprintf("%v hour%s ", hoursp, s)
		if days > 0 {
			return ss + "ago"
		}
	}

	minutes := int(since.Minutes())
	minutesp := minutes
	if hours > 0 {
		minutesp %= hours * 60
	}
	if minutes > 0 {
		s := ""
		if minutes > 1 {
			s = "s"
		}
		ss += fmt.Sprintf("%v minute%s ago", minutesp, s)
		return ss
	}

	sec := int(since.Seconds())
	if minutes != 0 {
		sec %= minutes
	}
	return fmt.Sprintf("%v seconds ago", sec)
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
