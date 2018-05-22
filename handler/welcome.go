package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

type WelcomeHandler struct {
	ss sessions.Store
}

func NewWelcomeHandler(ss sessions.Store) *WelcomeHandler {
	return &WelcomeHandler{
		ss: ss,
	}
}

func (h *WelcomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ses := getUserID(w, r, h.ss)
	if !ses {
		renderError(w, "Well that isn't good as the first thing you see after signing up... We had an error. Soz. Looks like you don't have a session... Try logging in?", ses)
		return
	}

	u, err := dao.ReadUserByID(context.Background(), userID)
	if err != nil {
		renderError(w, "Well that isn't good as the first thing you see after signing up... I can't read your user. Try again, or try logging in again?", ses)
		return
	}
	err = tmpl.GetTemplate("welcome").Execute(w, domain.BasePage{
		HasSession:  ses,
		SessionUser: u.User,
	})
	if err != nil {
		log.Printf("Template failed: %v", err)
		renderError(w, "Well that isn't good as the first thing you see after signing up... Something is wrong with the page template. Probably just go to the home page. The welcome page is boring anyway.", ses)
		return
	}
}
