package domain

import (
	"log"
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
	Details      string `datastore:",noindex"`
	RegisterDate time.Time
	Avatar       string `datastore:",noindex"`
	Follows      []string

	PostedCount     int
	PostedSentiment float64

	PostAttemptCount     int
	PostAttemptSentiment float64

	TotalSentimentEMA float64

	AngryBanExpire    time.Time
	AngryBanThreshold float64 `datastore:",noindex"`
	AngryBanCount     int
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

func (u User) AvatarLarge() string {
	a := u.AvatarUrl()

	if strings.Contains(a, "google") {
		return a + "?sz=500"
	}

	return a
}

func (u User) NameTitle() string {
	return strings.Title(u.Name)
}

func (u User) HappyPercent() int {

	sentiment := u.OverallSentiment()
	if sentiment > 0 {
		return int(sentiment * 100)
	}

	return 0
}

func (u User) AngryPercent() int {
	sentiment := u.OverallSentiment()
	if sentiment < 0 {
		return 100 + int(sentiment*100)
	}

	return 0
}

func (u User) OverallSentiment() float64 {
	return u.TotalSentimentEMA
}

func (u User) CalculatSentimentEMA(new float64) float64 {
	log.Printf("User %#v", u)

	if u.PostAttemptCount+u.PostedCount == 0 {
		log.Printf("returning new")
		return new
	}
	return ((new - u.TotalSentimentEMA) * 0.2) + u.TotalSentimentEMA
}

type UserPage struct {
	*BasePage
	User    User
	Posts   []PostWithUser
	Follows bool
	Sub     string
	ReplyTo string
}
