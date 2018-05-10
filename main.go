package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/handler"
	"github.com/jonomacd/forcedhappyness/site/statik"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

//go:generate statik -src=static/
func main() {

	dao.Init()
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
	http.Handle("/post", handler.NewPostHandler(sessionStore))
	http.Handle("/login", handler.NewLoginHandler(sessionStore))
	http.Handle("/register", handler.NewRegisterHandler(sessionStore))
	http.Handle("/like", handler.NewLikeHandler(sessionStore))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(statik.StatikFS)))
}

func genTestData() {
	err := dao.CreateUser(context.Background(), &dao.User{
		User: domain.User{
			Email:        "test@example.com",
			ID:           "1",
			Name:         "Jono MacDougall",
			RegisterDate: time.Now(),
		},
	})
	if err != nil {
		log.Printf("user create failed: %v", err)
	}

	err = dao.CreatePost(context.Background(), dao.Post{
		Post: domain.Post{
			ID:     "1",
			Text:   "foo bar",
			Date:   time.Now(),
			UserID: "1",
		}})
	if err != nil {
		log.Printf("Post read failed: %v", err)
	}

	err = dao.CreatePost(context.Background(), dao.Post{
		Post: domain.Post{
			ID:     "2",
			Text:   "foasdfo bar",
			Date:   time.Now(),
			UserID: "1",
		}})
	if err != nil {
		log.Printf("Post read failed: %v", err)
	}
	err = dao.CreatePost(context.Background(), dao.Post{
		Post: domain.Post{
			ID:     "3",
			Text:   "fhhhhhh bar",
			Date:   time.Now(),
			UserID: "1",
		}})
	if err != nil {
		log.Printf("Post read failed: %v", err)
	}
	err = dao.CreatePost(context.Background(), dao.Post{
		Post: domain.Post{
			ID:     "4",
			Text:   "sdfgasd bar",
			Date:   time.Now(),
			UserID: "1",
		}})
	if err != nil {
		log.Printf("Post read failed: %v", err)
	}
}
