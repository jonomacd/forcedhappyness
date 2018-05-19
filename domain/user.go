package domain

import (
	"strings"
	"time"

	"github.com/eefret/gravatar"
	"github.com/eefret/gravatar/default_img"
)

type User struct {
	ID           string
	Name         string
	Username     string
	Email        string
	Details      string
	RegisterDate time.Time
	Avatar       string
	Follows      []string
}

func (u User) AvatarUrl() string {
	if u.Avatar != "" {
		return u.Avatar
	}
	g, _ := gravatar.New()
	g.SetDefaultImage(default_img.DefaultImage.HTTP_404)
	g.SetSize(200)
	return g.URLParse(u.Email)
}

func (u User) NameTitle() string {
	return strings.Title(u.Name)
}

type UserPage struct {
	*BasePage
	User    User
	Posts   []PostWithUser
	Follows bool
}
