package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jonomacd/forcedhappyness/site/certificate"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/handler"
	"github.com/jonomacd/forcedhappyness/site/sentiment"
	"github.com/jonomacd/forcedhappyness/site/statik"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

//go:generate statik -src=static/
func main() {

	init := flag.Bool("init", false, "Use on first run on a new server")
	flag.Parse()
	if *init {
		if err := certificate.Run("jonomacd@gmail.com"); err != nil {
			log.Printf("error getting cert: %v", err)
		}
	}

	sentiment.InitNLP()
	statik.Init()
	tmpl.MustInit()

	f, err := statik.StatikFS.Open("/indexes/index.yaml")
	if err != nil {
		panic(err)
	}
	bb, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	if err := dao.Init(bb); err != nil {
		log.Printf("error init dao: %v", err)
	}

	registerHandlers()

	http.ListenAndServe("0.0.0.0:80", nil)
}

func registerHandlers() {

	sessionStore := dao.NewSessionStore()

	http.Handle("/", handler.NewHomeFeedHandler(sessionStore))
	http.Handle("/u/", handler.NewSubHandler(sessionStore))
	http.Handle("/sub", handler.NewSubCRUDHandler(sessionStore))
	http.Handle("/post/", handler.NewPostHandler(sessionStore))
	http.Handle("/settings", handler.NewSettingsHandler(sessionStore))
	http.Handle("/user/", handler.NewUserHandler(sessionStore))
	http.Handle("/submit", handler.NewSubmitHandler(sessionStore))
	http.Handle("/reply/", handler.NewReplyHandler(sessionStore))
	http.Handle("/login", handler.NewLoginHandler(sessionStore))
	http.Handle("/register", handler.NewRegisterHandler(sessionStore))
	http.Handle("/register/google", handler.NewGoogleRegisterHandler(sessionStore))
	http.Handle("/like", handler.NewLikeHandler(sessionStore))
	http.Handle("/follow", handler.NewFollowHandler(sessionStore))
	http.Handle("/welcome", handler.NewWelcomeHandler(sessionStore))
	http.Handle("/search", handler.NewSearchHandler(sessionStore))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(statik.StatikFS)))
}
