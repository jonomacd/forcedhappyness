package main

import (
	"net/http"

	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/handler"
	"github.com/jonomacd/forcedhappyness/site/sentiment"
	"github.com/jonomacd/forcedhappyness/site/statik"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

//go:generate statik -src=static/
func main() {

	dao.Init()
	sentiment.InitNLP()
	statik.Init()
	tmpl.MustInit()
	//genTestData()

	registerHandlers()

	http.ListenAndServe(":9091", nil)
}

func registerHandlers() {

	sessionStore := dao.NewSessionStore()

	http.Handle("/", handler.NewHomeFeedHandler(sessionStore))
	http.Handle("/u/", handler.NewSubHandler(sessionStore))
	http.Handle("/post/", handler.NewPostHandler(sessionStore))
	http.Handle("/settings", handler.NewSettingsHandler(sessionStore))
	http.Handle("/user/", handler.NewUserHandler(sessionStore))
	http.Handle("/submit", handler.NewSubmitHandler(sessionStore))
	http.Handle("/reply/", handler.NewReplyHandler(sessionStore))
	http.Handle("/login", handler.NewLoginHandler(sessionStore))
	http.Handle("/register", handler.NewRegisterHandler(sessionStore))
	http.Handle("/like", handler.NewLikeHandler(sessionStore))
	http.Handle("/follow", handler.NewFollowHandler(sessionStore))
	http.Handle("/welcome", handler.NewWelcomeHandler(sessionStore))
	http.Handle("/search", handler.NewSearchHandler(sessionStore))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(statik.StatikFS)))
}
