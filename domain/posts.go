package domain

import "time"

type Post struct {
	ID         string
	Text       string
	Date       time.Time
	UserID     string
	Sub        string
	Reputation float64
	Likes      int64
}

type PostWithUser struct {
	Post     Post
	User     User
	HasLiked bool
}

func (post Post) FormattedDate() string {
	return post.Date.Format(time.RFC822)
}

type PageData struct {
	BasePage
	Posts []PostWithUser
}
