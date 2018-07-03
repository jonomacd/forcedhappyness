package handler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/events"
)

type LikeHandler struct {
	ss sessions.Store
}

func NewLikeHandler(ss sessions.Store) *LikeHandler {
	return &LikeHandler{
		ss: ss,
	}
}

func (h *LikeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.post(w, r)
	case "DELETE":
		h.delete(w, r)
	}
}

func (h *LikeHandler) delete(w http.ResponseWriter, r *http.Request) {
	userID, hasSession := getUserID(w, r, h.ss)

	if !hasSession {
		redirectLogin(w, r)
		return
	}
	postID := r.URL.Query().Get("postID")
	if err := dao.DeleteLike(context.Background(), userID, postID); err != nil {
		log.Printf("cannot parse form c %v", err)
		return
	}
}

func (h *LikeHandler) post(w http.ResponseWriter, r *http.Request) {
	userID, hasSession := getUserID(w, r, h.ss)
	if !hasSession {
		redirectLogin(w, r)
		return
	}

	bb, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("cannot parse form a %v", err)
		return
	}
	defer r.Body.Close()
	body := map[string]string{}
	err = json.Unmarshal(bb, &body)
	if err != nil {
		log.Printf("cannot unmarshal %v", err)
		return
	}

	postID := body["postID"]

	_, err = dao.ReadLike(context.Background(), userID, postID)
	if err == dao.ErrNotFound {
		events.EventLike(events.LikeEvent{
			LikedBy: userID,
			Post:    postID,
		})
		if err := dao.CreateLike(context.Background(), userID, postID); err != nil {
			log.Printf("cannot parse form b %v", err)
			return
		}
		bb, _ := json.Marshal(map[string]bool{
			"liked": true,
		})
		w.Write(bb)
		return
	}
	if err == nil {
		if err := dao.DeleteLike(context.Background(), userID, postID); err != nil {
			log.Printf("cannot parse form c %v", err)
			return
		}
		bb, _ := json.Marshal(map[string]bool{
			"liked": false,
		})
		w.Write(bb)
		return
	}

}
