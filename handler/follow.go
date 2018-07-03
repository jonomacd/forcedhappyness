package handler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jonomacd/forcedhappyness/site/events"

	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/dao"
)

type FollowHandler struct {
	ss sessions.Store
}

func NewFollowHandler(ss sessions.Store) *FollowHandler {
	return &FollowHandler{
		ss: ss,
	}
}

func (h *FollowHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.post(w, r)
	case "DELETE":
		h.delete(w, r)
	}
}

func (h *FollowHandler) delete(w http.ResponseWriter, r *http.Request) {
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

func (h *FollowHandler) post(w http.ResponseWriter, r *http.Request) {
	userID, hasSession := getUserID(w, r, h.ss)
	if !hasSession {
		redirectLogin(w, r)
		return
	}

	ctx := context.Background()
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

	followUserID := body["userID"]

	if followUserID == userID {
		log.Printf("You can't follow yourself %s", userID)
		return
	}

	u, err := dao.ReadUserByID(ctx, userID)
	if err != nil {
		log.Printf("Error reading user %s: %v", userID, err)
		return
	}
	follows := false
	for _, fid := range u.Follows {
		if fid == followUserID {
			follows = true
			break
		}
	}

	if follows {
		err := dao.DeleteFollowerFromUser(ctx, userID, followUserID)
		if err != nil {
			log.Printf("Cannot delete follows %s %s: %v", userID, followUserID, err)
			return
		}
	} else {
		events.EventFollow(events.FollowEvent{
			FollowBy: userID,
			Followed: followUserID,
		})
		err := dao.AddFollowerToUser(ctx, userID, followUserID)
		if err != nil {
			log.Printf("Cannot add follows %s %s: %v", userID, followUserID, err)
			return
		}
	}

	bb, _ = json.Marshal(map[string]bool{
		"follows": !follows,
	})
	w.Write(bb)
	return

}
